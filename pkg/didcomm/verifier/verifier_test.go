package verifier

import (
	"testing"

	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/presentproof/engine"
)

func TestServer_RequestPresentation(t *testing.T) {

}

type providerMock struct {
}

func (r *providerMock) Store() datastore.Store {
	return nil
}

func (r *providerMock) GetAriesContext() (*ariescontext.Provider, error) {
	return nil, nil
}

func (r *providerMock) GetPresentationEngineRegistry() (engine.PresentationRegistry, error) {
	return nil, nil
}
