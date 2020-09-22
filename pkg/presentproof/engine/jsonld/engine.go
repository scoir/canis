package jsonld

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/scoir/canis/pkg/presentproof"
)

type Engine struct {
	ctx      *ariescontext.Provider
	proofsup *presentproof.Supervisor
	subject  *didexchange.Connection
	offerID  string
}
