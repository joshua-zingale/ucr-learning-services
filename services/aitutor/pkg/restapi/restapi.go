package restapi

import (
	"fmt"
	"net/http"
	"strconv"
)

type ResourceId map[string]string

func (r *ResourceId) GetInt(resourceName string) (int, error) {
	return strconv.Atoi((*r)[resourceName])
}

type Resource interface {
	Name() string
	PluralName() string
	Parent() Resource
}

type ResourceContext struct {
	ResourceId ResourceId
}

func GetResourceIdRaw(req *http.Request, res Resource) ResourceId {
	resourceId := make(ResourceId)
	curr := res
	for curr != nil {
		name := curr.Name()
		if value := req.PathValue(name); value != "" {
			resourceId[name] = value
		}
		curr = curr.Parent()
	}
	return resourceId
}

type ResourceHandler func(http.ResponseWriter, *http.Request, *ResourceContext)

func HandleResource(mux *http.ServeMux, method string, resource Resource, handler ResourceHandler) {
	urlPath := buildResourcePath(resource)
	mux.HandleFunc(fmt.Sprintf("%s %s", method, urlPath), func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, &ResourceContext{
			ResourceId: GetResourceIdRaw(r, resource),
		})
	})
}

func HandleResourceCollection(mux *http.ServeMux, method string, resource Resource, handler ResourceHandler) {
	urlPath := buildCollectionPath(resource)
	mux.HandleFunc(fmt.Sprintf("%s %s", method, urlPath), func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, &ResourceContext{
			ResourceId: GetResourceIdRaw(r, resource),
		})
	})
}

func buildResourcePath(r Resource) string {
	path := fmt.Sprintf("%s/{%s}", buildCollectionPath(r), r.Name())
	return path
}

func buildCollectionPath(r Resource) string {
	path := "/" + r.PluralName()
	if r.Parent() != nil {
		path = fmt.Sprintf("%s/{%s}%s", buildCollectionPath(r.Parent()), r.Parent().Name(), path)
	}
	return path
}
