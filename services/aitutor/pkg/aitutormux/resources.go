package aitutormux

import "github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/restapi"

type DirectResource struct {
	ResourceName       string
	ResourcePluralName string
	ParentResource     restapi.Resource
}

func NewResource(name string, pluralName string, parent restapi.Resource) restapi.Resource {
	return &DirectResource{
		ResourceName:       name,
		ResourcePluralName: pluralName,
		ParentResource:     parent,
	}
}

func (d *DirectResource) Name() string {
	return d.ResourceName
}

func (d *DirectResource) PluralName() string {
	return d.ResourcePluralName
}

func (d *DirectResource) Parent() restapi.Resource {
	return d.ParentResource
}

var AGENT_RESOURCE = NewResource("agent", "agents", nil)

var CONVERSATION_RESOURCE = NewResource("conversation", "conversations", AGENT_RESOURCE)
