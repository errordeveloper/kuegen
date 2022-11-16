package compiler

import (
	"encoding/json"
	"fmt"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

var sharedCUEMutex = &sync.Mutex{
	// load.Instances is not thread-safe (https://github.com/cue-lang/cue/issues/1043#issuecomment-1016729326)
}

type Compiler struct {
	ctx   *cue.Context
	mutex *sync.Mutex
}

func NewCompiler() *Compiler {
	return &Compiler{
		ctx:   cuecontext.New(),
		mutex: sharedCUEMutex,
	}
}

func (c *Compiler) BuildAll(dir, input string) (cue.Value, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	loadedInstances := load.Instances([]string{input}, &load.Config{Dir: dir})
	for _, loadedInstance := range loadedInstances {
		if loadedInstance.Err != nil {
			return cue.Value{}, fmtCUEError(fmt.Sprintf("failed to load instances (dir: %q, input: %q)", dir, input), loadedInstance.Err)
		}
	}

	builtInstances, err := c.ctx.BuildInstances(loadedInstances)
	if err != nil {
		return cue.Value{}, fmtCUEError(fmt.Sprintf("failed to build instances (dir: %q, input: %q)", dir, input), err)
	}
	for _, builtInstance := range builtInstances {
		if err := builtInstance.Value().Validate(); err != nil {
			return cue.Value{}, fmtCUEError("validation failure:", err)
		}
	}
	if len(builtInstances) != 1 {
		return cue.Value{}, fmt.Errorf("unexpected: more then one instance built")
	}
	return builtInstances[0], nil
}

func (c *Compiler) MarshalValueJSON(v cue.Value) ([]byte, error) {
	return json.Marshal(v)
}

func fmtCUEError(desc string, err error) error {
	return fmt.Errorf("%s: %s", desc, errors.Details(err, nil))
}
