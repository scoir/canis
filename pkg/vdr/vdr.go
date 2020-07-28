package vdr

import (
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
)

type Registry interface {
	Resolve(did string, opts ...vdri.ResolveOpts) (*diddoc.Doc, error)
	Store(doc *diddoc.Doc) error
	Create(method string, opts ...vdri.DocOpts) (*diddoc.Doc, error)
	Close() error
}
