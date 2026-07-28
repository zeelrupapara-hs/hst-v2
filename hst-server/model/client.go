package model

type ClientType int32

const (
	ClientType_undefined  ClientType = 0
	ClientType_individual ClientType = 1
	ClientType_corporate  ClientType = 2
	ClientType_fund       ClientType = 3
)

// Enum value maps for ClientType.
var (
	ClientType_name = map[int32]string{
		0: "undefined",
		1: "individual",
		2: "corporate",
		3: "fund",
	}
	ClientType_value = map[string]int32{
		"undefined":  0,
		"individual": 1,
		"corporate":  2,
		"fund":       3,
	}
)

type ClientStatus int32

// the ladder steps by 100, MT5 leaves room between states
const (
	ClientStatus_unregistered            ClientStatus = 0
	ClientStatus_registered              ClientStatus = 100
	ClientStatus_notinterested           ClientStatus = 200
	ClientStatus_application_incompleted ClientStatus = 300
	ClientStatus_application_completed   ClientStatus = 400
	ClientStatus_application_information ClientStatus = 500
	ClientStatus_application_rejected    ClientStatus = 600
	ClientStatus_approved                ClientStatus = 700
	ClientStatus_funded                  ClientStatus = 800
	ClientStatus_active                  ClientStatus = 900
	ClientStatus_inactive                ClientStatus = 1000
	ClientStatus_suspended               ClientStatus = 1100
	ClientStatus_closed                  ClientStatus = 1200
	ClientStatus_terminated              ClientStatus = 1300
)

// Enum value maps for ClientStatus.
var (
	ClientStatus_name = map[int32]string{
		0:    "unregistered",
		100:  "registered",
		200:  "notinterested",
		300:  "application_incompleted",
		400:  "application_completed",
		500:  "application_information",
		600:  "application_rejected",
		700:  "approved",
		800:  "funded",
		900:  "active",
		1000: "inactive",
		1100: "suspended",
		1200: "closed",
		1300: "terminated",
	}
	ClientStatus_value = map[string]int32{
		"unregistered":            0,
		"registered":              100,
		"notinterested":           200,
		"application_incompleted": 300,
		"application_completed":   400,
		"application_information": 500,
		"application_rejected":    600,
		"approved":                700,
		"funded":                  800,
		"active":                  900,
		"inactive":                1000,
		"suspended":               1100,
		"closed":                  1200,
		"terminated":              1300,
	}
)

type KycStatus int32

const (
	KycStatus_undefined KycStatus = 0
	KycStatus_approved  KycStatus = 1
	KycStatus_declined  KycStatus = 2
)

// Enum value maps for KycStatus.
var (
	KycStatus_name = map[int32]string{
		0: "undefined",
		1: "approved",
		2: "declined",
	}
	KycStatus_value = map[string]int32{
		"undefined": 0,
		"approved":  1,
		"declined":  2,
	}
)

type Gender int32

const (
	Gender_unspecified Gender = 0
	Gender_male        Gender = 1
	Gender_female      Gender = 2
)

// Enum value maps for Gender.
var (
	Gender_name = map[int32]string{
		0: "unspecified",
		1: "male",
		2: "female",
	}
	Gender_value = map[string]int32{
		"unspecified": 0,
		"male":        1,
		"female":      2,
	}
)

type Employment int32

const (
	Employment_unemployed    Employment = 0
	Employment_employed      Employment = 1
	Employment_self_employed Employment = 2
	Employment_retired       Employment = 3
	Employment_student       Employment = 4
	Employment_other         Employment = 5
)

// Enum value maps for Employment.
var (
	Employment_name = map[int32]string{
		0: "unemployed",
		1: "employed",
		2: "self_employed",
		3: "retired",
		4: "student",
		5: "other",
	}
	Employment_value = map[string]int32{
		"unemployed":    0,
		"employed":      1,
		"self_employed": 2,
		"retired":       3,
		"student":       4,
		"other":         5,
	}
)

type ClientIndustry int32

const (
	ClientIndustry_none          ClientIndustry = 0
	ClientIndustry_agriculture   ClientIndustry = 1
	ClientIndustry_construction  ClientIndustry = 2
	ClientIndustry_management    ClientIndustry = 3
	ClientIndustry_communication ClientIndustry = 4
	ClientIndustry_education     ClientIndustry = 5
	ClientIndustry_government    ClientIndustry = 6
	ClientIndustry_healthcare    ClientIndustry = 7
	ClientIndustry_tourism       ClientIndustry = 8
	ClientIndustry_it            ClientIndustry = 9
	ClientIndustry_security      ClientIndustry = 10
	ClientIndustry_manufacturing ClientIndustry = 11
	ClientIndustry_marketing     ClientIndustry = 12
	ClientIndustry_science       ClientIndustry = 13
	ClientIndustry_engineering   ClientIndustry = 14
	ClientIndustry_transport     ClientIndustry = 15
	ClientIndustry_other         ClientIndustry = 16
)

// Enum value maps for ClientIndustry.
var (
	ClientIndustry_name = map[int32]string{
		0:  "none",
		1:  "agriculture",
		2:  "construction",
		3:  "management",
		4:  "communication",
		5:  "education",
		6:  "government",
		7:  "healthcare",
		8:  "tourism",
		9:  "it",
		10: "security",
		11: "manufacturing",
		12: "marketing",
		13: "science",
		14: "engineering",
		15: "transport",
		16: "other",
	}
	ClientIndustry_value = map[string]int32{
		"none":          0,
		"agriculture":   1,
		"construction":  2,
		"management":    3,
		"communication": 4,
		"education":     5,
		"government":    6,
		"healthcare":    7,
		"tourism":       8,
		"it":            9,
		"security":      10,
		"manufacturing": 11,
		"marketing":     12,
		"science":       13,
		"engineering":   14,
		"transport":     15,
		"other":         16,
	}
)

type EducationLevel int32

const (
	EducationLevel_none        EducationLevel = 0
	EducationLevel_high_school EducationLevel = 1
	EducationLevel_bachelor    EducationLevel = 2
	EducationLevel_master      EducationLevel = 3
	EducationLevel_phd         EducationLevel = 4
	EducationLevel_other       EducationLevel = 5
)

// Enum value maps for EducationLevel.
var (
	EducationLevel_name = map[int32]string{
		0: "none",
		1: "high_school",
		2: "bachelor",
		3: "master",
		4: "phd",
		5: "other",
	}
	EducationLevel_value = map[string]int32{
		"none":        0,
		"high_school": 1,
		"bachelor":    2,
		"master":      3,
		"phd":         4,
		"other":       5,
	}
)

type WealthSource int32

const (
	WealthSource_employment  WealthSource = 0
	WealthSource_savings     WealthSource = 1
	WealthSource_inheritance WealthSource = 2
	WealthSource_other       WealthSource = 3
)

// Enum value maps for WealthSource.
var (
	WealthSource_name = map[int32]string{
		0: "employment",
		1: "savings",
		2: "inheritance",
		3: "other",
	}
	WealthSource_value = map[string]int32{
		"employment":  0,
		"savings":     1,
		"inheritance": 2,
		"other":       3,
	}
)

type PreferredCommunication int32

const (
	PreferredCommunication_undefined PreferredCommunication = 0
	PreferredCommunication_email     PreferredCommunication = 1
	PreferredCommunication_phone     PreferredCommunication = 2
	PreferredCommunication_phone_sms PreferredCommunication = 3
	PreferredCommunication_messenger PreferredCommunication = 4
)

// Enum value maps for PreferredCommunication.
var (
	PreferredCommunication_name = map[int32]string{
		0: "undefined",
		1: "email",
		2: "phone",
		3: "phone_sms",
		4: "messenger",
	}
	PreferredCommunication_value = map[string]int32{
		"undefined": 0,
		"email":     1,
		"phone":     2,
		"phone_sms": 3,
		"messenger": 4,
	}
)

type TradingExperience int32

const (
	TradingExperience_less_1_year  TradingExperience = 0
	TradingExperience_1_3_year     TradingExperience = 1
	TradingExperience_above_3_year TradingExperience = 2
)

// Enum value maps for TradingExperience.
var (
	TradingExperience_name = map[int32]string{
		0: "less_1_year",
		1: "1_3_year",
		2: "above_3_year",
	}
	TradingExperience_value = map[string]int32{
		"less_1_year":  0,
		"1_3_year":     1,
		"above_3_year": 2,
	}
)

type ClientOrigin int32

const (
	ClientOrigin_manual      ClientOrigin = 0
	ClientOrigin_demo        ClientOrigin = 1
	ClientOrigin_contest     ClientOrigin = 2
	ClientOrigin_preliminary ClientOrigin = 3
	ClientOrigin_real        ClientOrigin = 4
)

// Enum value maps for ClientOrigin.
var (
	ClientOrigin_name = map[int32]string{
		0: "manual",
		1: "demo",
		2: "contest",
		3: "preliminary",
		4: "real",
	}
	ClientOrigin_value = map[string]int32{
		"manual":      0,
		"demo":        1,
		"contest":     2,
		"preliminary": 3,
		"real":        4,
	}
)

// Client is the KYC person or legal entity. One client owns many logins.
type Client struct {
	ClientId                  int64        `db:"client_id" json:"client_id"`
	ClientType                ClientType   `db:"client_type" json:"client_type"`
	ClientStatus              ClientStatus `db:"client_status" json:"client_status"`
	KycStatus                 KycStatus    `db:"kyc_status" json:"kyc_status"`
	AssignedManager           int64        `db:"assigned_manager" json:"assigned_manager"`
	ComplianceApprovedBy      int64        `db:"compliance_approved_by" json:"compliance_approved_by"`
	ComplianceClientCategory  string       `db:"compliance_client_category" json:"compliance_client_category"`
	ComplianceDateApproval    int64        `db:"compliance_date_approval" json:"compliance_date_approval"`
	ComplianceDateTermination int64        `db:"compliance_date_termination" json:"compliance_date_termination"`
	Comment                   string       `db:"comment" json:"comment"`

	LeadCampaign      string       `db:"lead_campaign" json:"lead_campaign"`
	LeadSource        string       `db:"lead_source" json:"lead_source"`
	Introducer        int64        `db:"introducer" json:"introducer"`
	ClientOrigin      ClientOrigin `db:"client_origin" json:"client_origin"`
	ClientOriginLogin int64        `db:"client_origin_login" json:"client_origin_login"`

	PersonTitle          string         `db:"person_title" json:"person_title"`
	PersonName           string         `db:"person_name" json:"person_name"`
	PersonMiddleName     string         `db:"person_middle_name" json:"person_middle_name"`
	PersonLastName       string         `db:"person_last_name" json:"person_last_name"`
	PersonBirthDate      int64          `db:"person_birth_date" json:"person_birth_date"`
	PersonCitizenship    string         `db:"person_citizenship" json:"person_citizenship"`
	PersonGender         Gender         `db:"person_gender" json:"person_gender"`
	PersonTaxId          string         `db:"person_tax_id" json:"person_tax_id"`
	PersonDocumentType   string         `db:"person_document_type" json:"person_document_type"`
	PersonDocumentNumber string         `db:"person_document_number" json:"person_document_number"`
	PersonDocumentDate   int64          `db:"person_document_date" json:"person_document_date"`
	PersonDocumentExtra  string         `db:"person_document_extra" json:"person_document_extra"`
	PersonEmployment     Employment     `db:"person_employment" json:"person_employment"`
	PersonIndustry       ClientIndustry `db:"person_industry" json:"person_industry"`
	PersonEducation      EducationLevel `db:"person_education" json:"person_education"`
	PersonWealthSource   WealthSource   `db:"person_wealth_source" json:"person_wealth_source"`
	PersonAnnualIncome   float64        `db:"person_annual_income" json:"person_annual_income"`
	PersonNetWorth       float64        `db:"person_net_worth" json:"person_net_worth"`
	PersonAnnualDeposit  float64        `db:"person_annual_deposit" json:"person_annual_deposit"`

	CompanyName             string `db:"company_name" json:"company_name"`
	CompanyRegNumber        string `db:"company_reg_number" json:"company_reg_number"`
	CompanyRegDate          string `db:"company_reg_date" json:"company_reg_date"`
	CompanyRegAuthority     string `db:"company_reg_authority" json:"company_reg_authority"`
	CompanyVat              string `db:"company_vat" json:"company_vat"`
	CompanyLei              string `db:"company_lei" json:"company_lei"`
	CompanyLicenseNumber    string `db:"company_license_number" json:"company_license_number"`
	CompanyLicenseAuthority string `db:"company_license_authority" json:"company_license_authority"`
	CompanyCountry          string `db:"company_country" json:"company_country"`
	CompanyAddress          string `db:"company_address" json:"company_address"`
	CompanyWebsite          string `db:"company_website" json:"company_website"`

	ContactPreferred      PreferredCommunication `db:"contact_preferred" json:"contact_preferred"`
	ContactLanguage       string                 `db:"contact_language" json:"contact_language"`
	ContactEmail          string                 `db:"contact_email" json:"contact_email"`
	ContactPhone          string                 `db:"contact_phone" json:"contact_phone"`
	ContactMessengers     string                 `db:"contact_messengers" json:"contact_messengers"`
	ContactSocialNetworks string                 `db:"contact_social_networks" json:"contact_social_networks"`
	ContactLastDate       int64                  `db:"contact_last_date" json:"contact_last_date"`

	AddressCountry  string `db:"address_country" json:"address_country"`
	AddressPostcode string `db:"address_postcode" json:"address_postcode"`
	AddressStreet   string `db:"address_street" json:"address_street"`
	AddressState    string `db:"address_state" json:"address_state"`
	AddressCity     string `db:"address_city" json:"address_city"`

	ExperienceFx      TradingExperience `db:"experience_fx" json:"experience_fx"`
	ExperienceCfd     TradingExperience `db:"experience_cfd" json:"experience_cfd"`
	ExperienceFutures TradingExperience `db:"experience_futures" json:"experience_futures"`
	ExperienceStocks  TradingExperience `db:"experience_stocks" json:"experience_stocks"`

	DateCreated  int64 `db:"date_created" json:"date_created"`
	DateModified int64 `db:"date_modified" json:"date_modified"`
}

func (Client) TableName() string { return "hst.clients" }
