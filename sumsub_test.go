package sumsub

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goware/urlx"
	"github.com/op/go-logging"
)

var sumsub *SumSub
var applicantID = os.Getenv("SUMSUB_APPLICANT")

func init() {
	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}[%{module} %{shortfile}] %{message}%{color:reset}`,
	))
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
}

func TestURL(t *testing.T) {
	u, _ := urlx.Parse(TestAddr)
	s := &SumSub{
		url: *u,
	}

	addr := s.URL("/resources/auth/login")
	t.Log(addr)
	if !strings.HasSuffix(addr, "/resources/auth/login") {
		t.Error("failed to prepare url")
	}
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(TestAddr, os.Getenv("SUMSUB_USER"), os.Getenv("SUMSUB_PASS"))
	if err != nil {
		t.Error(err)
	}

	if c.token == "" {
		t.Error("token is empty")
	}

	if c.tokenExpired.Before(time.Now()) {
		t.Error("token expired")
	}

	sumsub = c

	t.Log(c.token)
}

func TestCreateApplicant(t *testing.T) {
	a := Applicant{
		ExternalUserID: "testid",
		Info: ApplicantInfo{
			Country:   "GBR",
			FirstName: "test",
			LastName:  "test",
		},
		RequiredIdDocs: ApplicantRequiredIDDocs{
			DocSets: []ApplicantDoc{
				ApplicantDoc{
					IDDocSetType: IDDocSetType_SELFIE,
					Types:        []string{DocSetType_SELFIE},
				},
			},
		},
	}

	if err := sumsub.CreateApplicant(&a); err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("%+v", a)

	applicantID = a.ID
}

func TestAddDocument(t *testing.T) {
	f, err := os.Open("testdata/selfie.jpg")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	metadata := DocumentMetaData{
		IDDocType: IDDocSetType_SELFIE,
		Country:   "USA",
	}

	var v interface{}
	if err := sumsub.AddDocument(applicantID, metadata, f, v); err != nil {
		t.Error(err)
	}

	t.Log(v)
}

func TestGetApplicant(t *testing.T) {
	a, err := sumsub.GetApplicant(applicantID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if a.ID != applicantID {
		t.Error("ID not equal")
	}

	t.Log(a)
}

func TestGetApplicantStatus(t *testing.T) {
	status, err := sumsub.GetApplicantStatus(applicantID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if status.ApplicantID != applicantID {
		t.Error("ID not equal")
	}

	t.Log(status)
}

func TestApplicantComplete(t *testing.T) {
	err := sumsub.ApplicantComplete(applicantID, ApplicantCompleteRequest{
		ReviewAnswer: ReviewResultGREEN,
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	status, err := sumsub.GetApplicantStatus(applicantID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !status.IsCompleted() {
		t.Log("status is not completed")
	}

	t.Log(status.IsPass())
}
