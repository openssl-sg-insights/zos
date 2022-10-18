package provision

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

// Response interface for custom error responses
// you never need to implement this interface
// can only be returned by one of the methods in this
// module.
type Response interface {
	error
	state() gridtypes.ResultState
}

type response struct {
	s gridtypes.ResultState
	e error
}

func (r *response) Error() string {
	if err := r.e; err != nil {
		return err.Error()
	}

	return ""
}

func (r *response) Unwrap() error {
	return r.e
}

func (r *response) state() gridtypes.ResultState {
	return r.s
}

// Ok response. you normally don't need to return
// this from Manager methods. instead returning `nil` error
// is preferred.
func Ok() Response {
	return &response{s: gridtypes.StateOk}
}

// UnChanged is a special response status that states that an operation has failed
// but this did not affect the workload status. Usually during an update when the
// update could not carried out, but the workload is still running correctly with
// previous config
func UnChanged(cause error) Response {
	return &response{s: gridtypes.StateUnChanged, e: cause}
}

func Paused() Response {
	return &response{s: gridtypes.StatePaused, e: fmt.Errorf("paused")}
}

// Result interface it's mainly a marker interface
// that does not implement a functionality
type Result interface {
	result()
}

// Resulter is a result that can return a result object
type Resulter interface {
	Result() gridtypes.Result
}

// NoActionResult is a result that marks a no action was done by a provisioner
// this means that a provisioner found out the operation were
// already done which means last result is still correct.
type NoActionResult interface {
	Result
	noAction()
}

// StateResult is a result that holds a non okay result
// this can be an error or a state change
type StateResult interface {
	Result
	Resulter
}

type ObjectResult interface {
	Result
	Resulter
}

type noActionResult struct{}

var (
	_ Result         = noActionResult{}
	_ NoActionResult = noActionResult{}
)

func (u noActionResult) result()   {}
func (u noActionResult) noAction() {}

func NewUnchangedResult() Result {
	return noActionResult{}
}

type stateResult struct {
	err error
}

var (
	_ Result   = stateResult{}
	_ Resulter = stateResult{}
)

func NewStateResult(err error) Resulter {
	if err != nil {
		panic("state result require a non-nil error")
	}

	return stateResult{err}
}

func (e stateResult) result() {}
func (e stateResult) Result() (result gridtypes.Result) {
	result.Created = gridtypes.Timestamp(time.Now().Unix())

	result.Error = e.err.Error()
	state := gridtypes.StateError

	var resp *response
	if errors.As(e.err, &resp) {
		state = resp.state()
	}

	result.State = state

	return
}

type objectResult struct {
	obj interface{}
}

var (
	_ Result   = objectResult{}
	_ Resulter = objectResult{}
)

func NewObjectResult(obj interface{}) Resulter {
	return objectResult{obj}
}

func (u objectResult) result() {}
func (u objectResult) Result() (result gridtypes.Result) {
	result.Created = gridtypes.Timestamp(time.Now().Unix())
	br, err := json.Marshal(u.obj)
	if err != nil {
		result.State = gridtypes.StateError
		result.Error = err.Error()
	}

	result.Data = br
	return
}
