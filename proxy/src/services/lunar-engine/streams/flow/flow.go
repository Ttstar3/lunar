package streamflow

import streamtypes "lunar/engine/streams/types"

type Flow struct{}

func NewFlow() *Flow {
	return &Flow{}
}

func (f *Flow) AddProcessor(_ streamtypes.ProcessorMetaData) {}

func (f *Flow) Execute(_ streamtypes.APIStream) {}
