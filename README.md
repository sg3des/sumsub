# SumSub API client

API client for [SumSub](https://sumsub.com/)

### Install

```sh
go get github.com/sg3des/sumsub
```

### USage


```go
ssapi,err:= sumsub.NewClient(sumsub.Addr, "user", "pass") // or sumsub.TestAddr
if err != nil {...}


// create applicant
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

if err := ssapi.CreateApplicant(&a); err != nil {
	...
}


// upload document to the applicant
f, err := os.Open("filename")
if err != nil {...}

docData := DocumentMetaData{
	IDDocType: IDDocSetType_SELFIE,
	Country:   "USA",
}

if err := ssapi.AddDocument(a.ID, docData, f, nil); err != nil {
	...
}

// get applicant status
status, err := ssapi.GetApplicantStatus(a.ID)
if err != nil {...}

if status.IsPass() {
	// then applicant approve
}else{
	status.ReviewResult.ModerationComment  // contains reason
}

```