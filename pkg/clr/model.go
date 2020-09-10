package clr

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/util"
)

type Address struct{}
type CryptographicKey struct{}

type Profile struct {
	ID                 string            `json:"id,omitempty"`
	Type               string            `json:"type"`
	Address            *Address          `json:"address,omitempty"`
	Description        string            `json:"description,omitempty"`
	Email              string            `json:"email,omitempty"`
	Endorsements       []*Endorsement    `json:"endorsements,omitempty"`
	Image              string            `json:"image,omitempty"`
	PublicKey          *CryptographicKey `json:"publicKey,omitempty"`
	RevocationList     string            `json:"revocationList,omitempty"`
	SignedEndorsements []*CompactJWS     `json:"signedEndorsements,omitempty"`
	SourcedID          string            `json:"sourcedId,omitempty"`
	StudentID          string            `json:"studentId,omitempty"`
	Telephone          string            `json:"telephone,omitempty"`
	URL                string            `json:"url,omitempty"`
	Verification       *Verification     `json:"verification,omitempty"`
}

type Endorsement struct{}
type Evidence struct{}
type Result struct{}
type Verification struct{}

type Assertion struct {
	ID                 string                         `json:"id,omitempty"`
	Type               string                         `json:"type"`
	Achievement        *Achievement                   `json:"achievement,omitempty"`
	CreditsEarned      int                            `json:"creditsEarned,omitempty"`
	EndDate            *util.TimeWithTrailingZeroMsec `json:"endDate,omitEmpty"`
	Endorsements       []*Endorsement                 `json:"endorsements,omitempty"`
	Evidence           []*Evidence                    `json:"evidence,omitempty"`
	Image              string                         `json:"image,omitempty"`
	IssuedOn           *util.TimeWithTrailingZeroMsec `json:"issuedOn,omitempty"`
	LicenseNumber      string                         `json:"licenseNumber,omitempty"`
	Narative           string                         `json:"narative,omitempty"`
	Recipient          string                         `json:"recipient,omitempty"`
	Results            []*Result                      `json:"results,omitempty"`
	RevocationReason   string                         `json:"revocationReason,omitempty"`
	Revoked            bool                           `json:"revoked,omitempty"`
	Role               string                         `json:"role,omitempty"`
	SignedEndorsements *CompactJWS                    `json:"signedEndorsements,omitempty"`
	Source             *Profile                       `json:"source,omitempty"`
	StartDate          *util.TimeWithTrailingZeroMsec `json:"startDate,omitempty"`
	Term               string                         `json:"term,omitempty"`
	Verification       *Verification                  `json:"verification,omitempty"`
}

type Alignment struct{}
type Association struct{}
type Criteria struct{}
type ResultDescription struct{}
type CompactJWS struct{}

type Achievement struct {
	ID                 string               `json:"id,omitempty"`
	Type               string               `json:"type,omitempty"`
	AchievementType    string               `json:"achievementType"`
	Alignments         []*Alignment         `json:"alignments,omitempty"`
	Associations       []*Association       `json:"associations,omitempty"`
	CreditsAvailable   int                  `json:"creditsAvailable,omitempty"`
	Description        string               `json:"description,omitempty"`
	HumanCode          string               `json:"humanCode,omitempty"`
	Name               string               `json:"name"`
	FieldOfStudy       string               `json:"fieldOfStudy,omitempty"`
	Image              string               `json:"image,omitempty"`
	Issuer             *Profile             `json:"issuer,omitempty"`
	Level              string               `json:"level,omitempty"`
	Requirement        *Criteria            `json:"requirement,omitempty"`
	ResultDescriptions []*ResultDescription `json:"resultDescriptions,omitempty"`
	SignedEndorsements *CompactJWS          `json:"signedEndorsements,omitempty"`
	Specialization     string               `json:"specialization,omitempty"`
	Tags               []string             `json:"tags,omitempty"`
}

type CLR struct {
	Context      []string                       `json:"@context,omitempty"`
	ID           string                         `json:"id,omitempty"`
	Type         string                         `json:"type"`
	Learner      *Profile                       `json:"learner"`
	Publisher    *Profile                       `json:"publisher"`
	Partial      bool                           `json:"partial,omitempty"`
	Assertions   []*Assertion                   `json:"assertions,omitempty"`
	Achievements []*Achievement                 `json:"achievements,omitempty"`
	IssuedOn     *util.TimeWithTrailingZeroMsec `json:"issuedOn,omitempty"`
}
