package proxywasm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_proxyOnContextCreate(t *testing.T) {
	currentStateMux.Lock()
	defer currentStateMux.Unlock()

	var cnt int
	currentState = &state{
		rootContexts:      map[uint32]*rootContextState{},
		httpStreams:       map[uint32]HttpContext{},
		streams:           map[uint32]StreamContext{},
		contextIDToRootID: map[uint32]uint32{},
	}

	SetNewRootContext(func(contextID uint32) RootContext {
		cnt++
		return nil
	})

	proxyOnContextCreate(100, 0)
	require.Equal(t, 1, cnt)
	SetNewHttpContext(func(rootContextID, contextID uint32) HttpContext {
		cnt += 100
		return nil
	})
	proxyOnContextCreate(100, 100)
	require.Equal(t, 101, cnt)
	currentState.newHttpContext = nil

	SetNewStreamContext(func(rootContextID, contextID uint32) StreamContext {
		cnt += 1000
		return nil
	})
	proxyOnContextCreate(100, 100)
	require.Equal(t, 1101, cnt)
}

type lifecycleContext struct {
	DefaultRootContext
	DefaultHttpContext
	DefaultStreamContext
	onStreamDone, onHttpStreamDone, onVMDone bool
}

func (ctx *lifecycleContext) OnVMDone() bool {
	ctx.onVMDone = true
	return true
}

func (ctx *lifecycleContext) OnStreamDone() {
	ctx.onStreamDone = true
}

func (ctx *lifecycleContext) OnHttpStreamDone() {
	ctx.onHttpStreamDone = true
}

func Test_onDone(t *testing.T) {
	currentStateMux.Lock()
	defer currentStateMux.Unlock()

	currentState = &state{
		rootContexts: map[uint32]*rootContextState{},
		httpStreams:  map[uint32]HttpContext{},
		streams:      map[uint32]StreamContext{},
	}

	var id uint32 = 1
	ctx := &lifecycleContext{}
	currentState.httpStreams[id] = ctx
	proxyOnDone(id)
	assert.True(t, ctx.onHttpStreamDone)
	assert.Equal(t, id, currentState.activeContextID)

	id = 2
	ctx = &lifecycleContext{}
	currentState.streams[id] = ctx
	proxyOnDone(id)
	assert.True(t, ctx.onStreamDone)
	assert.Equal(t, id, currentState.activeContextID)
}
