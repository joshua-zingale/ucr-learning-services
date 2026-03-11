package agentmux

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	reflectjsonschema "github.com/invopop/jsonschema"
	"github.com/jackc/pgtype"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/database"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/functools"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/qapi"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/templates"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type AuthService interface {
	Authenticate(r *http.Request) (UserProfile, error)
}

type UserProfile struct {
	UserId     string
	UserGroups []string
}

type ChatMessage struct {
	MessageType database.MessageType
	Content     string
}

// Defines functionality for an AgentClass
type AgentClassDriver interface {
	Generate(ctx context.Context, config []byte, messages []ChatMessage) (string, error)
	GetJsonSchemaRaw(ctx context.Context) ([]byte, error)
}

type AgentClassDriverRegistry interface {
	GetFromId(id string) (AgentClassDriver, bool)
}

type AgentArenaConfig struct {
	Db                       database.Queries
	Auth                     AuthService
	AgentClassDriverRegistry AgentClassDriverRegistry
}

type agentArenaHandler struct {
	*AgentArenaConfig
}

type contextKey int

const (
	idKey contextKey = iota
)

var templ = templates.LoadTemplates()

func NewAiTutorMux(config *AgentArenaConfig) http.Handler {
	if config == nil {
		panic("config cannot be nil")
	}

	mux := http.NewServeMux()

	h := agentArenaHandler{
		AgentArenaConfig: config,
	}

	fs := http.FileServer(http.Dir(filepath.Join("web", "static")))

	mux.HandleFunc("GET /agents/{agentId}", qapi.QApi(qapi.QApiParams[UserProfile, int32, agentWithConfigAndSchema]{
		Auth:        h.authenticate,
		Read:        setIdByPathParam("agentId"),
		Authz:       h.authorizeAgentAbility(database.AgentAbilityTypeManage),
		Act:         h.fetchAgentWithConfigAndSchema,
		Render:      renderTemplate[agentWithConfigAndSchema](templ, "agent.html"),
		RenderError: renderErrorAsHtml,
	}))
	mux.HandleFunc("GET /conversations", qapi.QApi(qapi.QApiParams[UserProfile, struct{}, []database.GetUserConversationsRow]{
		Auth:        h.authenticate,
		Read:        nil,
		Authz:       nil,
		Act:         h.fetchUserConversations,
		Render:      renderTemplate[[]database.GetUserConversationsRow](templ, "conversations.html"),
		RenderError: renderErrorAsHtml,
	}))
	mux.HandleFunc("GET /conversations/{conversationId}", qapi.QApi(qapi.QApiParams[UserProfile, int32, database.GetConversationRow]{
		Auth:        h.authenticate,
		Read:        setIdByPathParam("conversationId"),
		Authz:       h.authorizeStartedConversation,
		Act:         h.fetchConversation,
		Render:      renderTemplate[database.GetConversationRow](templ, "conversation.html"),
		RenderError: renderErrorAsHtml,
	}))

	mux.Handle("GET /static/", http.StripPrefix("/static", fs))

	mux.HandleFunc("GET /api/agents/{agentId}", qapi.QApi(qapi.QApiParams[UserProfile, int32, database.GetAgentRow]{
		Auth:        h.authenticate,
		Read:        setIdByPathParam("agentId"),
		Authz:       h.authorizeAgentAbility(database.AgentAbilityTypeManage),
		Act:         h.fetchAgent,
		Render:      renderJson[database.GetAgentRow],
		RenderError: renderErrorAsJson,
	}))

	mux.HandleFunc("PATCH /api/agents/{agentId}", qapi.QApi(qapi.QApiParams[UserProfile, editAgentParams, struct{}]{
		Auth:        h.authenticate,
		Read:        readEditAgentParams,
		Authz:       mapReq(h.authorizeAgentAbility(database.AgentAbilityTypeManage), func(eap editAgentParams) int32 { return eap.AgentId }),
		Act:         h.editAgent,
		RenderError: renderErrorAsJson,
	}))
	mux.HandleFunc("POST /api/agents", qapi.QApi(qapi.QApiParams[UserProfile, createAgentWithManagerRequest, createAgentWithManagerResponse]{
		Auth:        h.authenticate,
		Read:        readJson[createAgentWithManagerRequest],
		Authz:       authorizeGroup[createAgentWithManagerRequest]("agentarena.agentcreator"),
		Act:         h.createAgentWithManager,
		Render:      renderJson[createAgentWithManagerResponse],
		RenderError: renderErrorAsJson,
	}))

	mux.HandleFunc("GET /api/me/conversations", qapi.QApi(qapi.QApiParams[UserProfile, struct{}, []database.GetUserConversationsRow]{
		Auth:        h.authenticate,
		Read:        nil,
		Authz:       nil,
		Act:         h.fetchUserConversations,
		Render:      renderJson[[]database.GetUserConversationsRow],
		RenderError: renderErrorAsJson,
	}))

	mux.HandleFunc("POST /api/me/conversations", qapi.QApi(qapi.QApiParams[UserProfile, createConversationRequest, database.CreateConversationWithInitialMessageRow]{
		Auth:        h.authenticate,
		Read:        readJson[createConversationRequest],
		Authz:       mapReq(h.authorizeAgentAbility(database.AgentAbilityTypeInteract), func(ccr createConversationRequest) int32 { return ccr.AgentID }),
		Act:         h.createConversation,
		Render:      renderJson[database.CreateConversationWithInitialMessageRow],
		RenderError: renderErrorAsJson,
	}))

	mux.HandleFunc("GET /api/conversations/{conversationId}/messages", qapi.QApi(qapi.QApiParams[UserProfile, int32, []database.GetConversationMessagesRow]{
		Auth:        h.authenticate,
		Read:        setIdByPathParam("conversationId"),
		Authz:       h.authorizeStartedConversation,
		Act:         h.fetchConversationMessages,
		Render:      renderJson[[]database.GetConversationMessagesRow],
		RenderError: renderErrorAsJson,
	}))

	mux.HandleFunc("POST /api/conversations/{conversationId}/messages", qapi.QApi(qapi.QApiParams[UserProfile, createMessageRequest, database.PostMessageToConversationRow]{
		Auth:        h.authenticate,
		Read:        readJsonWithAugment(setIdByPathParam("conversationId"), func(i int32, cmr *createMessageRequest) { cmr.ConversationId = i }),
		Authz:       mapReq(h.authorizeStartedConversation, func(cmr createMessageRequest) int32 { return cmr.ConversationId }),
		Act:         h.createMessage,
		Render:      renderJson[database.PostMessageToConversationRow],
		RenderError: renderErrorAsJson,
	}))

	return mux
}

func (ath *agentArenaHandler) fetchUserConversations(ctx context.Context, profile UserProfile, _ struct{}) ([]database.GetUserConversationsRow, error) {
	conversations, err := ath.Db.GetUserConversations(ctx, profile.UserId)
	if err != nil {
		return nil, &internalError{}
	}
	return conversations, nil
}

type createConversationRequest struct {
	MessageContent   string `json:"messageContent"`
	ConversationName string `json:"conversationName"`
	AgentID          int32  `json:"agentId"`
}

func (ath *agentArenaHandler) createConversation(ctx context.Context, profile UserProfile, req createConversationRequest) (database.CreateConversationWithInitialMessageRow, error) {
	createdData, err := ath.Db.CreateConversationWithInitialMessage(ctx, database.CreateConversationWithInitialMessageParams{
		MessageContent:   req.MessageContent,
		UserID:           profile.UserId,
		ConversationName: req.ConversationName,
		AgentID:          sql.NullInt32{Int32: req.AgentID, Valid: true},
	})
	if err != nil {
		return createdData, &internalError{err}
	}
	return createdData, nil

}

func (ath *agentArenaHandler) fetchAgent(ctx context.Context, _ UserProfile, agentId int32) (database.GetAgentRow, error) {

	agent, err := ath.Db.GetAgent(ctx, agentId)
	if err != nil {
		return agent, &notFoundError{}
	}

	return agent, nil
}

type createAgentWithManagerRequest struct {
	Name         string `json:"name"`
	AgentClassID string `json:"agentClassId"`
	Config       any    `json:"config"`
}

type createAgentWithManagerResponse struct {
	AgentId int32 `json:"agentId"`
}

func (ath *agentArenaHandler) createAgentWithManager(ctx context.Context, manager UserProfile, req createAgentWithManagerRequest) (createAgentWithManagerResponse, error) {
	var resp createAgentWithManagerResponse
	var config pgtype.JSONB
	if err := config.Set(req.Config); err != nil {
		return resp, &internalError{err}
	}
	driver, ok := ath.AgentClassDriverRegistry.GetFromId(req.AgentClassID)
	if !ok {
		return resp, fmt.Errorf("invalid agent class '%s'", req.AgentClassID)
	}

	schema, err := driver.GetJsonSchemaRaw(ctx)
	if err != nil {
		return resp, &internalError{err}
	}

	agentId, err := ath.Db.CreateAgentWithManagerAndInteractor(ctx, schema, database.CreateAgentWithManagerAndInteractor{
		Name:         req.Name,
		AgentClassID: req.AgentClassID,
		Config:       config,
		UserID:       manager.UserId,
	})
	if err != nil {
		return resp, &internalError{err}
	}

	resp.AgentId = agentId
	return resp, nil
}

type agentWithConfigAndSchema struct {
	AgentID      int32
	Name         string
	Config       string
	ConfigSchema string
}

func (ath *agentArenaHandler) fetchAgentWithConfigAndSchema(ctx context.Context, userProfile UserProfile, agentId int32) (agentWithConfigAndSchema, error) {
	var awcas agentWithConfigAndSchema

	agent, err := ath.fetchAgent(ctx, userProfile, agentId)
	if err != nil {
		return awcas, err
	}

	driver, ok := ath.AgentClassDriverRegistry.GetFromId(agent.AgentClassID)
	if !ok {
		return awcas, &internalError{fmt.Errorf("missing config schema for agent class %s", agent.AgentClassID)}
	}

	configSchema, err := driver.GetJsonSchemaRaw(ctx)
	if err != nil {
		return awcas, &internalError{err}
	}

	awcas.AgentID = agent.AgentID
	awcas.Name = agent.Name
	awcas.Config = string(agent.Config.Bytes)
	awcas.ConfigSchema = string(configSchema)

	return awcas, nil
}

type editAgentParams struct {
	AgentId int32  `json:"agentId"`
	Config  []byte `json:"config"`
}

func readEditAgentParams(r *http.Request) (editAgentParams, error) {
	var p editAgentParams
	id, err := setIdByPathParam("agentId")(r)
	if err != nil {
		return p, err
	}
	config, err := io.ReadAll(r.Body)
	if err != nil {
		return p, &internalError{err}
	}
	p.AgentId = id
	p.Config = config
	return p, nil
}

func (ath *agentArenaHandler) editAgent(ctx context.Context, _ UserProfile, p editAgentParams) (struct{}, error) {

	agent, err := ath.Db.GetAgent(ctx, p.AgentId)
	if err != nil {
		return struct{}{}, &notFoundError{}
	}

	driver, ok := ath.AgentClassDriverRegistry.GetFromId(agent.AgentClassID)
	if !ok {
		return struct{}{}, &internalError{fmt.Errorf("missing driver for agent class with id %s", agent.AgentClassID)}
	}

	schema, err := driver.GetJsonSchemaRaw(ctx)
	if err != nil {
		return struct{}{}, &internalError{err}
	}

	if err = ath.Db.UpdateAgentConfig(ctx, agent, p.Config, schema); err != nil {
		switch err.(type) {
		case *jsonschema.ValidationError:
			return struct{}{}, err
		default:
			return struct{}{}, &internalError{err}
		}

	}

	return struct{}{}, nil
}

func (ath *agentArenaHandler) fetchConversationMessages(ctx context.Context, _ UserProfile, conversationId int32) ([]database.GetConversationMessagesRow, error) {

	messages, err := ath.Db.GetConversationMessages(ctx, conversationId)
	if err != nil {
		return messages, &internalError{err}
	}

	return messages, nil
}

func (ath *agentArenaHandler) fetchConversation(ctx context.Context, _ UserProfile, conversationId int32) (database.GetConversationRow, error) {

	conversation, err := ath.Db.GetConversation(ctx, conversationId)
	if err != nil {
		return conversation, &internalError{err}
	}

	return conversation, nil
}

type createMessageRequest struct {
	ConversationId int32                `json:"conversationId,omitempty"`
	MessageType    database.MessageType `json:"messageType"`
	Content        string               `json:"content,omitempty"`
}

func (ath *agentArenaHandler) createMessage(ctx context.Context, userProfile UserProfile, req createMessageRequest) (database.PostMessageToConversationRow, error) {

	var newMessage database.PostMessageToConversationRow

	switch req.MessageType {
	case database.MessageTypeUser:
		if strings.TrimSpace(req.Content) == "" {
			return newMessage, fmt.Errorf("cannot post empty message")
		}
		userMessage, err := ath.Db.PostMessageToConversation(ctx, database.PostMessageToConversationParams{
			ConversationID: req.ConversationId,
			Content:        req.Content,
			MessageType:    database.MessageTypeUser,
			UserID:         sql.NullString{String: userProfile.UserId, Valid: true},
		})

		if err != nil {
			return userMessage, &internalError{err}
		}

		return userMessage, nil
	case database.MessageTypeAgent:
		agentConfig, err := ath.Db.GetAgentConfigFromConversationId(ctx, req.ConversationId)
		if err != nil {
			return newMessage, &internalError{fmt.Errorf("getting config %d %w", req.ConversationId, err)}
		}

		agentDriver, ok := ath.AgentClassDriverRegistry.GetFromId(agentConfig.AgentClassID)
		if !ok {
			return newMessage, &internalError{fmt.Errorf("getting agent driver: %w", err)}
		}

		messages, err := ath.Db.GetConversationMessages(ctx, req.ConversationId)
		if err != nil {
			return newMessage, &internalError{fmt.Errorf("getting conversation messages: %w", err)}
		}

		chatMessages := functools.Map(messages, func(m database.GetConversationMessagesRow) ChatMessage {
			return ChatMessage{
				MessageType: m.MessageType,
				Content:     m.Content,
			}
		})
		agentMessageContent, err := agentDriver.Generate(ctx, agentConfig.Config.Bytes, chatMessages)
		if err != nil {
			return newMessage, &internalError{err}
		}

		agentMessage, err := ath.Db.PostMessageToConversation(ctx, database.PostMessageToConversationParams{
			ConversationID: req.ConversationId,
			Content:        agentMessageContent,
			MessageType:    database.MessageTypeAgent,
			AgentID:        sql.NullInt32{Int32: agentConfig.AgentID, Valid: true},
		})

		if err != nil {
			return agentMessage, &internalError{err}
		}

		return agentMessage, nil
	default:
		return newMessage, fmt.Errorf("bad message type '%s'", req.MessageType)
	}

}

func (ath *agentArenaHandler) authenticate(r *http.Request) (UserProfile, error) {
	var profile UserProfile
	var err error
	profile, err = ath.Auth.Authenticate(r)
	if err != nil {
		return profile, &unauthorizedError{}
	}
	return profile, nil
}

func authorizeGroup[Any any](group string) qapi.AuthzFunction[UserProfile, Any] {
	return func(_ context.Context, profile UserProfile, _ Any) error {
		for _, userGroup := range profile.UserGroups {
			if group == userGroup {
				return nil
			}
		}

		return &forbiddenError{}
	}
}

func readJsonWithAugment[Aug, Req any](agumentFetcher qapi.ReadRequestFunction[Aug], augmenter func(Aug, *Req)) qapi.ReadRequestFunction[Req] {
	return func(r *http.Request) (Req, error) {
		var req Req

		aug, err := agumentFetcher(r)
		if err != nil {
			return req, err
		}

		req, err = readJson[Req](r)
		if err != nil {
			return req, err
		}

		augmenter(aug, &req)

		return req, nil
	}
}

var schemaCache sync.Map

func readJson[Req any](r *http.Request) (Req, error) {
	var req Req

	t := reflect.TypeOf(req)
	compiled, err := getOrCompileSchema(t)
	if err != nil {
		return req, fmt.Errorf("schema compilation error: %w", err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return req, fmt.Errorf("reading request body: %w", err)
	}
	defer r.Body.Close()

	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return req, fmt.Errorf("malformed json: %w", err)
	}

	if err := compiled.Validate(v); err != nil {
		return req, err
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("unmarshaling to struct: %w", err)
	}

	return req, nil
}

func getOrCompileSchema(t reflect.Type) (*jsonschema.Schema, error) {
	if cached, ok := schemaCache.Load(t); ok {
		return cached.(*jsonschema.Schema), nil
	}

	reflector := &reflectjsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schemaObj := reflector.ReflectFromType(t)
	schemaBytes, _ := json.Marshal(schemaObj)

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return nil, err
	}

	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, err
	}

	schemaCache.Store(t, compiled)
	return compiled, nil
}

func mapReq[Any, Req1, Req2 any](fn qapi.AuthzFunction[Any, Req1], mapper func(Req2) Req1) qapi.AuthzFunction[Any, Req2] {
	return func(ctx context.Context, a Any, r Req2) error {
		return fn(ctx, a, mapper(r))
	}
}

func (ath *agentArenaHandler) authorizeAgentAbility(ability database.AgentAbilityType) qapi.AuthzFunction[UserProfile, int32] {
	return func(ctx context.Context, profile UserProfile, agentId int32) error {
		hasPermission, err := ath.Db.HasAgentPermission(ctx, database.HasAgentPermissionParams{
			UserID:   profile.UserId,
			GroupIds: profile.UserGroups,
			AgentID:  agentId,
			Ability:  ability,
		})
		if err != nil {
			return &internalError{err}
		}
		if hasPermission {
			return nil
		}
		return &forbiddenError{}
	}

}

func (ath *agentArenaHandler) authorizeStartedConversation(ctx context.Context, profile UserProfile, conversationId int32) error {
	if hasPermission, err := ath.Db.StartedConversation(ctx, database.StartedConversationParams{
		ConversationID: int32(conversationId),
		UserID:         profile.UserId,
	}); err != nil || !hasPermission {
		return &forbiddenError{}
	}

	return nil
}

func setIdByPathParam(pathParamName string) qapi.ReadRequestFunction[int32] {
	return func(r *http.Request) (int32, error) {
		id, err := strconv.ParseInt(r.PathValue(pathParamName), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid integer ID as path parameter")
		}
		return int32(id), nil
	}

}

type agentIdJson struct {
	AgentId int32 `json:"agentId"`
}

func renderJson[T any](_ context.Context, w http.ResponseWriter, data T) {
	respondJson(w, data)
}

type errorMessage struct {
	Error string `json:"error"`
}

func renderErrorAsJson(_ context.Context, w http.ResponseWriter, data error) {
	w.Header().Set("Content-Type", "application/json")

	var msg any

	switch data.(type) {
	case *jsonschema.ValidationError:
		w.WriteHeader(http.StatusUnprocessableEntity)
		msg = map[string]any{
			"error":  "validation_failed",
			"fields": extractLeafErrors(data),
		}
	case *internalError:
		w.WriteHeader(http.StatusInternalServerError)
		msg = errorMessage{"Internal Error"}
		log.Print(data.Error())
	case *forbiddenError:
		w.WriteHeader(http.StatusForbidden)
		msg = errorMessage{data.Error()}
	case *unauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)
		msg = errorMessage{data.Error()}
	default:
		w.WriteHeader(http.StatusBadRequest)
		msg = errorMessage{data.Error()}
	}
	respondJson(w, msg)
}

func renderErrorAsHtml(_ context.Context, w http.ResponseWriter, data error) {
	w.Header().Set("Content-Type", "text/html")

	switch data.(type) {
	case *internalError:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Error"))
		log.Print(data.Error())
	case *forbiddenError:
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(data.Error()))
	case *unauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(data.Error()))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(data.Error()))
	}
}

func extractLeafErrors(err error) map[string]string {
	flat := make(map[string]string)

	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {

			key := e.InstanceLocation
			if key == "" {
				key = "root"
			}
			flat[key] = e.Message
		} else {
			for _, cause := range e.Causes {
				walk(cause)
			}
		}
	}

	if ve, ok := err.(*jsonschema.ValidationError); ok {
		walk(ve)
	}
	return flat
}

func renderTemplate[T any](tmpl *template.Template, name string) qapi.RenderFunction[T] {
	return func(_ context.Context, w http.ResponseWriter, data T) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := tmpl.ExecuteTemplate(w, name, data)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

func respondJson[T any](w http.ResponseWriter, v T) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(v)
}

type unauthorizedError struct{}

func (ue *unauthorizedError) Error() string {
	return "Unauthorized"
}

type forbiddenError struct{}

func (fe *forbiddenError) Error() string {
	return "Forbidden"
}

type internalError struct {
	InternalError error
}

func (ie *internalError) Error() string {

	if ie.InternalError != nil {
		return fmt.Sprintf("Internal Error: %s", ie.InternalError.Error())
	}
	return "Internal Error"
}

type notFoundError struct{}

func (nfe *notFoundError) Error() string {
	return "Not Found"
}
