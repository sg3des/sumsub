package sumsub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"time"

	"github.com/goware/urlx"
	"github.com/imroc/req"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("SUMSUB")

	// development address
	TestAddr = "https://test-api.sumsub.com"

	// production address
	Addr = "https://api.sumsub.com"

	// token lifetime is 7 days, recomended renew token earlier
	tokenLifetime = time.Hour * 150
)

// SumSub
type SumSub struct {
	url  url.URL
	user string
	pass string

	token        string
	tokenExpired time.Time
}

// NewClient to sumsub server, prepare sumsub struct instance and obtain token
func NewClient(addr, user, pass string) (*SumSub, error) {
	u, err := urlx.ParseWithDefaultScheme(addr, "https")
	if err != nil {
		return nil, err
	}

	s := &SumSub{
		url:  *u,
		user: user,
		pass: pass,
	}

	token, err := s.Authentication(user, pass)
	if err != nil {
		return s, fmt.Errorf("token not recieved: %v", err)
	}

	s.token = token
	s.tokenExpired = time.Now().Add(tokenLifetime)

	return s, nil
}

func (s *SumSub) URL(urlpath ...string) string {
	s.url.Path = path.Join(urlpath...)
	return s.url.String()
}

func (s *SumSub) authHeader() req.Header {
	return req.Header{
		"Authorization": "Bearer " + s.token,
	}
}

// Authentication request to obtain `token`
// POST /resources/auth/login
// https://developers.sumsub.com/#authentication
func (s *SumSub) Authentication(user, pass string) (token string, err error) {
	basic := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	header := req.Header{
		"Authorization": "Basic " + basic,
	}
	resp, err := req.Post(s.URL("/resources/auth/login"), header)
	if err != nil {
		return "", err
	} else if r := resp.Response(); r.StatusCode != 200 {
		return "", errors.New(r.Status)
	}

	var aResp authResp
	if err := resp.ToJSON(&aResp); err != nil {
		return "", err
	}

	if aResp.Status != "ok" || aResp.Payload == "" {
		return "", errors.New("failed token obtain")
	}

	return aResp.Payload, nil
}

type authResp struct {
	Status  string
	Payload string
}

type Error struct {
	Description   string
	Code          int
	CorrelationId string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Description)
}

func handleResponse(resp *req.Resp, err error) error {
	if err != nil {
		return err
	}

	if r := resp.Response(); r.StatusCode >= 400 {
		err := &Error{
			Code: r.StatusCode,
		}

		resp.ToJSON(err)

		return err
	}

	return nil
}

//
// Applicants API
// https://developers.sumsub.com/#applicants-api
//

type Applicant struct {

	// request
	ExternalUserID string   `json:"externalUserId"`
	SourceKey      string   `json:"sourceKey,omitempty"`
	Email          string   `json:"email,omitempty"`
	Lang           string   `json:"lang,omitempty"`
	Metadata       []string `json:"metadata,omitempty"`

	Info           ApplicantInfo           `json:"info"`
	RequiredIdDocs ApplicantRequiredIDDocs `json:"requiredIdDocs"`

	// response
	ID           string `json:"id,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	InspectionID string `json:"inspectionId,omitempty"`
	JobID        string `json:"jobId,omitempty"`
	Env          string `json:"env,omitempty"`

	Review struct {
		CreateDate             string      `json:"createDate"`
		ReviewResult           interface{} `json:"reviewResult"`
		ReviewStatus           string      `json:"reviewStatus"`
		NotificationFailureCnt int         `json:"notificationFailureCnt"`
	} `json:"review,omitempty"`
}

type ApplicantInfo struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`

	Gender       string `json:"gender,omitempty"`
	DateOfBirth  string `json:"dob,omitempty"`
	PlaceOfBirth string `json:"placeOfBirth,omitempty"`

	Country string `json:"country,omitempty"`
	Phone   string `json:"phone,omitempty"`

	Addresses []Address `json:"addresses,omitempty"`
}

type Address struct {
	Country   string `json:"country,omitempty"`
	PostCode  string `json:"postCode,omitempty"`
	Town      string `json:"town,omitempty"`
	Street    string `json:"street,omitempty"`
	SubStreet string `json:"subStreet,omitempty"`
	State     string `json:"state,omitempty"`
}

type ApplicantRequiredIDDocs struct {
	Country           string   `json:"country,omitempty"`
	IncludedCountries []string `json:"includedCountries,omitempty"`
	ExcludedCountries []string `json:"excludedCountries,omitempty"`

	DocSets []ApplicantDoc `json:"docSets"`
}

const (
	IDDocSetType_IDENTITY           = "IDENTITY"
	IDDocSetType_IDENTITY2          = "IDENTITY2"
	IDDocSetType_SELFIE             = "SELFIE"
	IDDocSetType_SELFIE2            = "SELFIE2"
	IDDocSetType_PROOF_OF_RESIDENCE = "PROOF_OF_RESIDENCE"
	IDDocSetType_PAYMENT_METHODS    = "PAYMENT_METHODS"
)

const (
	DocSetType_ID_CARD                          = "ID_CARD"
	DocSetType_PASSPORT                         = "PASSPORT"
	DocSetType_DRIVERS                          = "DRIVERS"
	DocSetType_BANK_CARD                        = "BANK_CARD"
	DocSetType_UTILITY_BILL                     = "UTILITY_BILL"
	DocSetType_BANK_STATEMENT                   = "BANK_STATEMENT"
	DocSetType_SNILS                            = "SNILS"
	DocSetType_SELFIE                           = "SELFIE"
	DocSetType_VIDEO_SELFIE                     = "VIDEO_SELFIE"
	DocSetType_PROFILE_IMAGE                    = "PROFILE_IMAGE"
	DocSetType_ID_DOC_PHOTO                     = "ID_DOC_PHOTO"
	DocSetType_AGREEMENT                        = "AGREEMENT"
	DocSetType_CONTRACT                         = "CONTRACT"
	DocSetType_RESIDENCE_PERMIT                 = "RESIDENCE_PERMIT"
	DocSetType_EMPLOYMENT_CERTIFICATE           = "EMPLOYMENT_CERTIFICATE"
	DocSetType_DRIVERS_TRANSLATION              = "DRIVERS_TRANSLATION"
	DocSetType_INVESTOR_DOC                     = "INVESTOR_DOC"
	DocSetType_VEHICLE_REGISTRATION_CERTIFICATE = "VEHICLE_REGISTRATION_CERTIFICATE"
	DocSetType_INCOME_SOURCE                    = "INCOME SOURCE"
	DocSetType_OTHER                            = "OTHER"
)

const (
	DocSetSubTypeFront = "FRONT_SIDE"
	DocSetSubTypeBack  = "BACK_SIDE"
)

type ApplicantDoc struct {
	IDDocSetType string   `json:"idDocSetType"`
	Types        []string `json:"types"`
	SubTypes     []string `json:"subTypes,omitempty"`
	Fields       []string `json:"fields,omitempty"`
	ImageIDs     []string `json:"imageIds,omitempty"`
}

// CreateApplicant entity representing one physical person. It may have several
// ID documents attached, like an ID card or a passport. Many additional photos
// of different documents can be attached to the same applicant.
// POST /resources/applicants
// https://developers.sumsub.com/#creating-an-applicant
func (s *SumSub) CreateApplicant(a *Applicant) error {
	resp, err := req.Post(s.URL("resources/applicants"), s.authHeader(), req.BodyJSON(a))
	if err := handleResponse(resp, err); err != nil {
		return err
	}

	return resp.ToJSON(&a)
}

type DocumentMetaData struct {
	IDDocType    string `json:"idDocType"`
	IDDocSubType string `json:"idDocSubType,omitempty"`
	Country      string `json:"country"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	MiddleName   string `json:"middleName,omitempty"`
	IssuedDate   string `json:"issuedDate,omitempty"`
	ValidUntil   string `json:"validUntil,omitempty"`
	Number       string `json:"number,omitempty"`
	DateOfBirth  string `json:"dob,omitempty"`
	PlaceOfBirth string `json:"placeOfBirth,omitempty"`
}

// AddDocument to applicant, it required metadata with description of the file
func (s *SumSub) AddDocument(id string, metadata DocumentMetaData, file io.Reader, v interface{}) error {
	var bufMetdata bytes.Buffer
	json.NewEncoder(&bufMetdata).Encode(metadata)

	reqMetdata := req.FileUpload{
		FieldName: "metadata",
		File:      ioutil.NopCloser(&bufMetdata),
	}

	reqContent := req.FileUpload{
		FieldName: "content",
		File:      ioutil.NopCloser(file),
	}

	resp, err := req.Post(s.URL("resources/applicants/"+id+"/info/idDoc"), s.authHeader(), reqMetdata, reqContent)
	if err := handleResponse(resp, err); err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	return resp.ToJSON(&v)
}

type applicantsList struct {
	List struct {
		Items      []Applicant
		TotalItems int
	}
}

func (s *SumSub) GetApplicant(id string) (a Applicant, err error) {
	resp, err := req.Get(s.URL("resources/applicants/"+id), s.authHeader())
	if err := handleResponse(resp, err); err != nil {
		return a, err
	}

	var list applicantsList
	if err := resp.ToJSON(&list); err != nil {
		return a, err
	}
	if len(list.List.Items) == 0 {
		return a, errors.New("applicant not found")
	}

	return list.List.Items[0], nil
}

type ApplicantStatus struct {
	ID           string `json:"id"`
	InspectionID string `json:"inspectionId"`
	ApplicantID  string `json:"applicantId"`
	JobID        string `json:"jobId"`

	CreateDate string `json:"createDate"`
	StartDate  string `json:"startDate"`

	ReviewResult ReviewResult `json:"reviewResult"`

	ReviewStatus           string `json:"reviewStatus"`
	NotificationFailureCnt int    `json:"notificationFailureCnt"`
}

func (status ApplicantStatus) IsCompleted() bool {
	return status.ReviewStatus == ReviewStatusCompleted ||
		status.ReviewStatus == ReviewStatusCompletedSent ||
		status.ReviewStatus == ReviewStatusCompletedSetFailure
}

func (status ApplicantStatus) IsPass() (string, bool) {
	return status.ReviewResult.ModerationComment, status.ReviewResult.ReviewAnswer == ReviewResultGREEN
}

type ReviewResult struct {
	ModerationComment string   `json:"moderationComment"`
	ClientComment     string   `json:"clientComment"`
	ReviewAnswer      string   `json:"reviewAnswer"`
	RejectLabels      []string `json:"rejectLabels"`
	ReviewRejectType  string   `json:"reviewRejectType"`
	CustomTouch       bool     `json:"customTouch"`
}

const (
	ReviewStatusInit                = "init"
	ReviewStatusPending             = "pending"
	ReviewStatusQueued              = "queued"
	ReviewStatusCompleted           = "completed"
	ReviewStatusCompletedSent       = "completedSent"
	ReviewStatusCompletedSetFailure = "completedSentFailure"
)

const (
	ReviewResultRED   = "RED"
	ReviewResultGREEN = "GREEN"
)

func (s *SumSub) GetApplicantStatus(id string) (a ApplicantStatus, err error) {
	resp, err := req.Get(s.URL("resources/applicants/"+id+"/status"), s.authHeader())
	if err := handleResponse(resp, err); err != nil {
		return a, err
	}

	err = resp.ToJSON(&a)
	return
}

type ApplicantCompleteRequest struct {
	ReviewAnswer     string   `json:"reviewAnswer"`
	RejectLabels     []string `json:"rejectLabels"`
	ReviewRejectType string   `json:"reviewRejectType,omitempty"`
}

func (s *SumSub) ApplicantComplete(id string, data ApplicantCompleteRequest) error {
	resp, err := req.Post(s.URL("resources/applicants/"+id+"/status/testCompleted"), s.authHeader(), req.BodyJSON(data))
	return handleResponse(resp, err)
}
