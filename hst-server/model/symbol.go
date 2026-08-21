package model

type SymbolIndustry int32

const (
	SymbolIndustry_undefined                    SymbolIndustry = 0
	SymbolIndustry_agricultural_inputs          SymbolIndustry = 1
	SymbolIndustry_aluminium                    SymbolIndustry = 2
	SymbolIndustry_building_materials           SymbolIndustry = 3
	SymbolIndustry_chemicals                    SymbolIndustry = 4
	SymbolIndustry_coking_coal                  SymbolIndustry = 5
	SymbolIndustry_copper                       SymbolIndustry = 6
	SymbolIndustry_gold                         SymbolIndustry = 7
	SymbolIndustry_lumber_wood                  SymbolIndustry = 8
	SymbolIndustry_industrial_metals            SymbolIndustry = 9
	SymbolIndustry_precious_metals              SymbolIndustry = 10
	SymbolIndustry_paper                        SymbolIndustry = 11
	SymbolIndustry_silver                       SymbolIndustry = 12
	SymbolIndustry_specialty_chemicals          SymbolIndustry = 13
	SymbolIndustry_steel                        SymbolIndustry = 14
	SymbolIndustry_advertising                  SymbolIndustry = 15
	SymbolIndustry_broadcasting                 SymbolIndustry = 16
	SymbolIndustry_gaming_multimedia            SymbolIndustry = 17
	SymbolIndustry_entertainment                SymbolIndustry = 18
	SymbolIndustry_internet_content             SymbolIndustry = 19
	SymbolIndustry_publishing                   SymbolIndustry = 20
	SymbolIndustry_telecom                      SymbolIndustry = 21
	SymbolIndustry_apparel_manufacturing        SymbolIndustry = 22
	SymbolIndustry_apparel_retail               SymbolIndustry = 23
	SymbolIndustry_auto_manufacturers           SymbolIndustry = 24
	SymbolIndustry_auto_parts                   SymbolIndustry = 25
	SymbolIndustry_auto_dealership              SymbolIndustry = 26
	SymbolIndustry_department_stores            SymbolIndustry = 27
	SymbolIndustry_footwear_accessories         SymbolIndustry = 28
	SymbolIndustry_furnishings                  SymbolIndustry = 29
	SymbolIndustry_gambling                     SymbolIndustry = 30
	SymbolIndustry_home_improv_retail           SymbolIndustry = 31
	SymbolIndustry_internet_retail              SymbolIndustry = 32
	SymbolIndustry_leisure                      SymbolIndustry = 33
	SymbolIndustry_lodging                      SymbolIndustry = 34
	SymbolIndustry_luxury_goods                 SymbolIndustry = 35
	SymbolIndustry_packaging_containers         SymbolIndustry = 36
	SymbolIndustry_personal_services            SymbolIndustry = 37
	SymbolIndustry_recreational_vehicles        SymbolIndustry = 38
	SymbolIndustry_resident_construction        SymbolIndustry = 39
	SymbolIndustry_resorts_casinos              SymbolIndustry = 40
	SymbolIndustry_restaurants                  SymbolIndustry = 41
	SymbolIndustry_specialty_retail             SymbolIndustry = 42
	SymbolIndustry_textile_manufacturing        SymbolIndustry = 43
	SymbolIndustry_travel_services              SymbolIndustry = 44
	SymbolIndustry_beverages_brewers            SymbolIndustry = 45
	SymbolIndustry_beverages_non_alco           SymbolIndustry = 46
	SymbolIndustry_beverages_wineries           SymbolIndustry = 47
	SymbolIndustry_confectioners                SymbolIndustry = 48
	SymbolIndustry_discount_stores              SymbolIndustry = 49
	SymbolIndustry_education_trainig            SymbolIndustry = 50
	SymbolIndustry_farm_products                SymbolIndustry = 51
	SymbolIndustry_food_distribution            SymbolIndustry = 52
	SymbolIndustry_grocery_stores               SymbolIndustry = 53
	SymbolIndustry_household_products           SymbolIndustry = 54
	SymbolIndustry_packaged_foods               SymbolIndustry = 55
	SymbolIndustry_tobacco                      SymbolIndustry = 56
	SymbolIndustry_oil_gas_drilling             SymbolIndustry = 57
	SymbolIndustry_oil_gas_ep                   SymbolIndustry = 58
	SymbolIndustry_oil_gas_equipment            SymbolIndustry = 59
	SymbolIndustry_oil_gas_integrated           SymbolIndustry = 60
	SymbolIndustry_oil_gas_midstream            SymbolIndustry = 61
	SymbolIndustry_oil_gas_refining             SymbolIndustry = 62
	SymbolIndustry_thermal_coal                 SymbolIndustry = 63
	SymbolIndustry_uranium                      SymbolIndustry = 64
	SymbolIndustry_exchange_traded_fund         SymbolIndustry = 65
	SymbolIndustry_assets_management            SymbolIndustry = 66
	SymbolIndustry_banks_diversified            SymbolIndustry = 67
	SymbolIndustry_banks_regional               SymbolIndustry = 68
	SymbolIndustry_capital_markets              SymbolIndustry = 69
	SymbolIndustry_close_end_fund_debt          SymbolIndustry = 70
	SymbolIndustry_close_end_fund_equity        SymbolIndustry = 71
	SymbolIndustry_close_end_fund_foreign       SymbolIndustry = 72
	SymbolIndustry_credit_services              SymbolIndustry = 73
	SymbolIndustry_financial_conglomerate       SymbolIndustry = 74
	SymbolIndustry_financial_data_exchange      SymbolIndustry = 75
	SymbolIndustry_insurance_brokers            SymbolIndustry = 76
	SymbolIndustry_insurance_diversified        SymbolIndustry = 77
	SymbolIndustry_insurance_life               SymbolIndustry = 78
	SymbolIndustry_insurance_property           SymbolIndustry = 79
	SymbolIndustry_insurance_reinsurance        SymbolIndustry = 80
	SymbolIndustry_insurance_specialty          SymbolIndustry = 81
	SymbolIndustry_mortgage_finance             SymbolIndustry = 82
	SymbolIndustry_shell_companies              SymbolIndustry = 83
	SymbolIndustry_biotechnology                SymbolIndustry = 84
	SymbolIndustry_diagnostics_research         SymbolIndustry = 85
	SymbolIndustry_drugs_manufacturers          SymbolIndustry = 86
	SymbolIndustry_drugs_manufacturers_spec     SymbolIndustry = 87
	SymbolIndustry_healthcare_plans             SymbolIndustry = 88
	SymbolIndustry_health_information           SymbolIndustry = 89
	SymbolIndustry_medical_facilities           SymbolIndustry = 90
	SymbolIndustry_medical_devices              SymbolIndustry = 91
	SymbolIndustry_medical_distribution         SymbolIndustry = 92
	SymbolIndustry_medical_instruments          SymbolIndustry = 93
	SymbolIndustry_pharm_retailers              SymbolIndustry = 94
	SymbolIndustry_aerospace_defense            SymbolIndustry = 95
	SymbolIndustry_airlines                     SymbolIndustry = 96
	SymbolIndustry_airports_services            SymbolIndustry = 97
	SymbolIndustry_building_products            SymbolIndustry = 98
	SymbolIndustry_business_equipment           SymbolIndustry = 99
	SymbolIndustry_conglomerates                SymbolIndustry = 100
	SymbolIndustry_consulting_services          SymbolIndustry = 101
	SymbolIndustry_electrical_equipment         SymbolIndustry = 102
	SymbolIndustry_engineering_construction     SymbolIndustry = 103
	SymbolIndustry_farm_heavy_machinery         SymbolIndustry = 104
	SymbolIndustry_industrial_distribution      SymbolIndustry = 105
	SymbolIndustry_infrastructure_operations    SymbolIndustry = 106
	SymbolIndustry_freight_logistics            SymbolIndustry = 107
	SymbolIndustry_marine_shipping              SymbolIndustry = 108
	SymbolIndustry_metal_fabrication            SymbolIndustry = 109
	SymbolIndustry_pollution_control            SymbolIndustry = 110
	SymbolIndustry_railroads                    SymbolIndustry = 111
	SymbolIndustry_rental_leasing               SymbolIndustry = 112
	SymbolIndustry_security_protection          SymbolIndustry = 113
	SymbolIndustry_speality_business_services   SymbolIndustry = 114
	SymbolIndustry_speality_machinery           SymbolIndustry = 115
	SymbolIndustry_stuffing_employment          SymbolIndustry = 116
	SymbolIndustry_tools_accessories            SymbolIndustry = 117
	SymbolIndustry_trucking                     SymbolIndustry = 118
	SymbolIndustry_waste_management             SymbolIndustry = 119
	SymbolIndustry_real_estate_development      SymbolIndustry = 120
	SymbolIndustry_real_estate_diversified      SymbolIndustry = 121
	SymbolIndustry_real_estate_services         SymbolIndustry = 122
	SymbolIndustry_reit_diversified             SymbolIndustry = 123
	SymbolIndustry_reit_healtcare               SymbolIndustry = 124
	SymbolIndustry_reit_hotel_motel             SymbolIndustry = 125
	SymbolIndustry_reit_industrial              SymbolIndustry = 126
	SymbolIndustry_reit_mortage                 SymbolIndustry = 127
	SymbolIndustry_reit_office                  SymbolIndustry = 128
	SymbolIndustry_reit_residental              SymbolIndustry = 129
	SymbolIndustry_reit_retail                  SymbolIndustry = 130
	SymbolIndustry_reit_speciality              SymbolIndustry = 131
	SymbolIndustry_communication_equipment      SymbolIndustry = 132
	SymbolIndustry_computer_hardware            SymbolIndustry = 133
	SymbolIndustry_consumer_electronics         SymbolIndustry = 134
	SymbolIndustry_electronic_components        SymbolIndustry = 135
	SymbolIndustry_electronic_distribution      SymbolIndustry = 136
	SymbolIndustry_it_services                  SymbolIndustry = 137
	SymbolIndustry_scientific_instruments       SymbolIndustry = 138
	SymbolIndustry_semiconductor_equipment      SymbolIndustry = 139
	SymbolIndustry_semiconductors               SymbolIndustry = 140
	SymbolIndustry_software_application         SymbolIndustry = 141
	SymbolIndustry_software_infrastructure      SymbolIndustry = 142
	SymbolIndustry_solar                        SymbolIndustry = 143
	SymbolIndustry_utilities_diversified        SymbolIndustry = 144
	SymbolIndustry_utilities_powerproducers     SymbolIndustry = 145
	SymbolIndustry_utilities_renewable          SymbolIndustry = 146
	SymbolIndustry_utilities_regulated_electric SymbolIndustry = 147
	SymbolIndustry_utilities_regulated_gas      SymbolIndustry = 148
	SymbolIndustry_utilities_regulated_water    SymbolIndustry = 149
)

var (
	SymbolIndustry_name = map[int32]string{
		0:   "undefined",
		1:   "agricultural_inputs",
		2:   "aluminium",
		3:   "building_materials",
		4:   "chemicals",
		5:   "coking_coal",
		6:   "copper",
		7:   "gold",
		8:   "lumber_wood",
		9:   "industrial_metals",
		10:  "precious_metals",
		11:  "paper",
		12:  "silver",
		13:  "specialty_chemicals",
		14:  "steel",
		15:  "advertising",
		16:  "broadcasting",
		17:  "gaming_multimedia",
		18:  "entertainment",
		19:  "internet_content",
		20:  "publishing",
		21:  "telecom",
		22:  "apparel_manufacturing",
		23:  "apparel_retail",
		24:  "auto_manufacturers",
		25:  "auto_parts",
		26:  "auto_dealership",
		27:  "department_stores",
		28:  "footwear_accessories",
		29:  "furnishings",
		30:  "gambling",
		31:  "home_improv_retail",
		32:  "internet_retail",
		33:  "leisure",
		34:  "lodging",
		35:  "luxury_goods",
		36:  "packaging_containers",
		37:  "personal_services",
		38:  "recreational_vehicles",
		39:  "resident_construction",
		40:  "resorts_casinos",
		41:  "restaurants",
		42:  "specialty_retail",
		43:  "textile_manufacturing",
		44:  "travel_services",
		45:  "beverages_brewers",
		46:  "beverages_non_alco",
		47:  "beverages_wineries",
		48:  "confectioners",
		49:  "discount_stores",
		50:  "education_trainig",
		51:  "farm_products",
		52:  "food_distribution",
		53:  "grocery_stores",
		54:  "household_products",
		55:  "packaged_foods",
		56:  "tobacco",
		57:  "oil_gas_drilling",
		58:  "oil_gas_ep",
		59:  "oil_gas_equipment",
		60:  "oil_gas_integrated",
		61:  "oil_gas_midstream",
		62:  "oil_gas_refining",
		63:  "thermal_coal",
		64:  "uranium",
		65:  "exchange_traded_fund",
		66:  "assets_management",
		67:  "banks_diversified",
		68:  "banks_regional",
		69:  "capital_markets",
		70:  "close_end_fund_debt",
		71:  "close_end_fund_equity",
		72:  "close_end_fund_foreign",
		73:  "credit_services",
		74:  "financial_conglomerate",
		75:  "financial_data_exchange",
		76:  "insurance_brokers",
		77:  "insurance_diversified",
		78:  "insurance_life",
		79:  "insurance_property",
		80:  "insurance_reinsurance",
		81:  "insurance_specialty",
		82:  "mortgage_finance",
		83:  "shell_companies",
		84:  "biotechnology",
		85:  "diagnostics_research",
		86:  "drugs_manufacturers",
		87:  "drugs_manufacturers_spec",
		88:  "healthcare_plans",
		89:  "health_information",
		90:  "medical_facilities",
		91:  "medical_devices",
		92:  "medical_distribution",
		93:  "medical_instruments",
		94:  "pharm_retailers",
		95:  "aerospace_defense",
		96:  "airlines",
		97:  "airports_services",
		98:  "building_products",
		99:  "business_equipment",
		100: "conglomerates",
		101: "consulting_services",
		102: "electrical_equipment",
		103: "engineering_construction",
		104: "farm_heavy_machinery",
		105: "industrial_distribution",
		106: "infrastructure_operations",
		107: "freight_logistics",
		108: "marine_shipping",
		109: "metal_fabrication",
		110: "pollution_control",
		111: "railroads",
		112: "rental_leasing",
		113: "security_protection",
		114: "speality_business_services",
		115: "speality_machinery",
		116: "stuffing_employment",
		117: "tools_accessories",
		118: "trucking",
		119: "waste_management",
		120: "real_estate_development",
		121: "real_estate_diversified",
		122: "real_estate_services",
		123: "reit_diversified",
		124: "reit_healtcare",
		125: "reit_hotel_motel",
		126: "reit_industrial",
		127: "reit_mortage",
		128: "reit_office",
		129: "reit_residental",
		130: "reit_retail",
		131: "reit_speciality",
		132: "communication_equipment",
		133: "computer_hardware",
		134: "consumer_electronics",
		135: "electronic_components",
		136: "electronic_distribution",
		137: "it_services",
		138: "scientific_instruments",
		139: "semiconductor_equipment",
		140: "semiconductors",
		141: "software_application",
		142: "software_infrastructure",
		143: "solar",
		144: "utilities_diversified",
		145: "utilities_powerproducers",
		146: "utilities_renewable",
		147: "utilities_regulated_electric",
		148: "utilities_regulated_gas",
		149: "utilities_regulated_water",
	}
	SymbolIndustry_label = map[int32]string{
		0:   "Undefined",
		1:   "Agricultural inputs",
		2:   "Aluminium",
		3:   "Building materials",
		4:   "Chemicals",
		5:   "Coking coal",
		6:   "Copper",
		7:   "Gold",
		8:   "Lumber and wood production",
		9:   "Other industrial metals and mining",
		10:  "Other precious metals and mining",
		11:  "Paper and paper products",
		12:  "Silver",
		13:  "Specialty chemicals",
		14:  "Steel",
		15:  "Advertising agencies",
		16:  "Broadcasting",
		17:  "Electronic gaming and multimedia",
		18:  "Entertainment",
		19:  "Internet content and information",
		20:  "Publishing",
		21:  "Telecom services",
		22:  "Apparel manufacturing",
		23:  "Apparel retail",
		24:  "Auto manufacturers",
		25:  "Auto parts",
		26:  "Auto and truck dealerships",
		27:  "Department stores",
		28:  "Footwear and accessories",
		29:  "Furnishing, fixtures and appliances",
		30:  "Gambling",
		31:  "Home improvement retail",
		32:  "Internet retail",
		33:  "Leisure",
		34:  "Lodging",
		35:  "Luxury goods",
		36:  "Packaging and containers",
		37:  "Personal services",
		38:  "Recreational vehicles",
		39:  "Residential construction",
		40:  "Resorts and casinos",
		41:  "Restaurants",
		42:  "Specialty retail",
		43:  "Textile manufacturing",
		44:  "Travel services",
		45:  "Beverages - Brewers",
		46:  "Beverages - Non-alcoholic",
		47:  "Beverages - Wineries and distilleries",
		48:  "Confectioners",
		49:  "Discount stores",
		50:  "Education and training services",
		51:  "Farm products",
		52:  "Food distribution",
		53:  "Grocery stores",
		54:  "Household and personal products",
		55:  "Packaged foods",
		56:  "Tobacco",
		57:  "Oil and gas drilling",
		58:  "Oil and gas extraction and processing",
		59:  "Oil and gas equipment and services",
		60:  "Oil and gas integrated",
		61:  "Oil and gas midstream",
		62:  "Oil and gas refining and marketing",
		63:  "Thermal coal",
		64:  "Uranium",
		65:  "Exchange traded fund",
		66:  "Assets management",
		67:  "Banks - Diversified",
		68:  "Banks - Regional",
		69:  "Capital markets",
		70:  "Closed-End fund - Debt",
		71:  "Closed-end fund - Equity",
		72:  "Closed-end fund - Foreign",
		73:  "Credit services",
		74:  "Financial conglomerates",
		75:  "Financial data and stock exchange",
		76:  "Insurance brokers",
		77:  "Insurance - Diversified",
		78:  "Insurance - Life",
		79:  "Insurance - Property and casualty",
		80:  "Insurance - Reinsurance",
		81:  "Insurance - Specialty",
		82:  "Mortgage finance",
		83:  "Shell companies",
		84:  "Biotechnology",
		85:  "Diagnostics and research",
		86:  "Drugs manufacturers - general",
		87:  "Drugs manufacturers - Specialty and generic",
		88:  "Healthcare plans",
		89:  "Health information services",
		90:  "Medical care facilities",
		91:  "Medical devices",
		92:  "Medical distribution",
		93:  "Medical instruments and supplies",
		94:  "Pharmaceutical retailers",
		95:  "Aerospace and defense",
		96:  "Airlines",
		97:  "Airports and air services",
		98:  "Building products and equipment",
		99:  "Business equipment and supplies",
		100: "Conglomerates",
		101: "Consulting services",
		102: "Electrical equipment and parts",
		103: "Engineering and construction",
		104: "Farm and heavy construction machinery",
		105: "Industrial distribution",
		106: "Infrastructure operations",
		107: "Integrated freight and logistics",
		108: "Marine shipping",
		109: "Metal fabrication",
		110: "Pollution and treatment controls",
		111: "Railroads",
		112: "Rental and leasing services",
		113: "Security and protection services",
		114: "Specialty business services",
		115: "Specialty industrial machinery",
		116: "Stuffing and employment services",
		117: "Tools and accessories",
		118: "Trucking",
		119: "Waste management",
		120: "Real estate - Development",
		121: "Real estate - Diversified",
		122: "Real estate services",
		123: "REIT - Diversified",
		124: "REIT - Healthcase facilities",
		125: "REIT - Hotel and motel",
		126: "REIT - Industrial",
		127: "REIT - Mortgage",
		128: "REIT - Office",
		129: "REIT - Residential",
		130: "REIT - Retail",
		131: "REIT - Specialty",
		132: "Communication equipment",
		133: "Computer hardware",
		134: "Consumer electronics",
		135: "Electronic components",
		136: "Electronics and computer distribution",
		137: "Information technology services",
		138: "Scientific and technical instruments",
		139: "Semiconductor equipment and materials",
		140: "Semiconductors",
		141: "Software - Application",
		142: "Software - Infrastructure",
		143: "Solar",
		144: "Utilities - Diversified",
		145: "Utilities - Independent power producers",
		146: "Utilities - Renewable",
		147: "Utilities - Regulated electric",
		148: "Utilities - Regulated gas",
		149: "Utilities - Regulated water",
	}
)

// SymbolIndustrySector returns the hst-v2 sector for an industry id.
func SymbolIndustrySector(industry int32) SymbolSector {
	switch {
	case industry >= 1 && industry <= 14:
		return SymbolSector(1)
	case industry >= 15 && industry <= 21:
		return SymbolSector(2)
	case industry >= 22 && industry <= 44:
		return SymbolSector(3)
	case industry >= 45 && industry <= 56:
		return SymbolSector(4)
	case industry >= 57 && industry <= 64:
		return SymbolSector(5)
	case industry >= 65 && industry <= 83:
		return SymbolSector(6)
	case industry >= 84 && industry <= 95:
		return SymbolSector(7)
	case industry >= 95 && industry <= 119:
		return SymbolSector(8)
	case industry >= 120 && industry <= 131:
		return SymbolSector(9)
	case industry >= 132 && industry <= 143:
		return SymbolSector(10)
	case industry >= 144 && industry <= 149:
		return SymbolSector(11)
	default:
		return SymbolSector_undefined
	}
}

// Symbol-domain enums.
var (
	CalcMode_name = map[int32]string{
		0:  "forex",
		1:  "futures",
		2:  "cfd",
		3:  "cfd_index",
		4:  "cfd_leverage",
		5:  "forex_no_leverage",
		32: "exch_stocks",
		33: "exch_futures",
		34: "exch_forts",
		35: "exch_options",
		36: "exch_options_margin",
		37: "exch_bonds",
		64: "serv_collateral",
	}
	CalcMode_value = map[string]int32{
		"forex":               0,
		"futures":             1,
		"cfd":                 2,
		"cfd_index":           3,
		"cfd_leverage":        4,
		"forex_no_leverage":   5,
		"exch_stocks":         32,
		"exch_futures":        33,
		"exch_forts":          34,
		"exch_options":        35,
		"exch_options_margin": 36,
		"exch_bonds":          37,
		"serv_collateral":     64,
	}
)

var (
	TradeMode_name = map[int32]string{
		0: "disabled",
		1: "long_only",
		2: "short_only",
		3: "close_only",
		4: "full",
	}
	TradeMode_value = map[string]int32{
		"disabled":   0,
		"long_only":  1,
		"short_only": 2,
		"close_only": 3,
		"full":       4,
	}
)

var (
	ExecMode_name = map[int32]string{
		0: "request",
		1: "instant",
		2: "market",
		3: "exchange",
	}
	ExecMode_value = map[string]int32{
		"request":  0,
		"instant":  1,
		"market":   2,
		"exchange": 3,
	}
)

type GTCMode int32

const (
	GTCMode_gtc            GTCMode = 0
	GTCMode_daily          GTCMode = 1
	GTCMode_daily_no_stops GTCMode = 2
)

var (
	GTCMode_name = map[int32]string{
		0: "gtc",
		1: "daily",
		2: "daily_no_stops",
	}
	GTCMode_value = map[string]int32{
		"gtc":            0,
		"daily":          1,
		"daily_no_stops": 2,
	}
)

type FillingFlags int32

const (
	FillingFlags_none FillingFlags = 0
	FillingFlags_fok  FillingFlags = 1
	FillingFlags_ioc  FillingFlags = 2
	FillingFlags_boc  FillingFlags = 4
)

var (
	FillingFlags_name = map[int32]string{
		0: "none",
		1: "fok",
		2: "ioc",
		4: "boc",
	}
	FillingFlags_value = map[string]int32{
		"none": 0,
		"fok":  1,
		"ioc":  2,
		"boc":  4,
	}
)

type ExpirationFlags int32

const (
	ExpirationFlags_none          ExpirationFlags = 0
	ExpirationFlags_gtc           ExpirationFlags = 1
	ExpirationFlags_day           ExpirationFlags = 2
	ExpirationFlags_specified     ExpirationFlags = 4
	ExpirationFlags_specified_day ExpirationFlags = 8
)

var (
	ExpirationFlags_name = map[int32]string{
		0: "none",
		1: "gtc",
		2: "day",
		4: "specified",
		8: "specified_day",
	}
	ExpirationFlags_value = map[string]int32{
		"none":          0,
		"gtc":           1,
		"day":           2,
		"specified":     4,
		"specified_day": 8,
	}
)

type OrderFlags int32

const (
	OrderFlags_none       OrderFlags = 0
	OrderFlags_market     OrderFlags = 1
	OrderFlags_limit      OrderFlags = 2
	OrderFlags_stop       OrderFlags = 4
	OrderFlags_stop_limit OrderFlags = 8
	OrderFlags_sl         OrderFlags = 16
	OrderFlags_tp         OrderFlags = 32
	OrderFlags_closeby    OrderFlags = 64

	OrderFlags_all OrderFlags = 127
)

var (
	OrderFlags_name = map[int32]string{
		0:   "none",
		1:   "market",
		2:   "limit",
		4:   "stop",
		8:   "stop_limit",
		16:  "sl",
		32:  "tp",
		64:  "closeby",
		127: "all",
	}
	OrderFlags_value = map[string]int32{
		"none":       0,
		"market":     1,
		"limit":      2,
		"stop":       4,
		"stop_limit": 8,
		"sl":         16,
		"tp":         32,
		"closeby":    64,
		"all":        127,
	}
)

type SwapMode int32

const (
	SwapMode_disabled              SwapMode = 0
	SwapMode_by_points             SwapMode = 1
	SwapMode_by_symbol_currency    SwapMode = 2
	SwapMode_by_margin_currency    SwapMode = 3
	SwapMode_by_group_currency     SwapMode = 4
	SwapMode_by_interest_current   SwapMode = 5
	SwapMode_by_interest_open      SwapMode = 6
	SwapMode_reopen_by_close_price SwapMode = 7
	SwapMode_reopen_by_bid         SwapMode = 8
	SwapMode_by_profit_currency    SwapMode = 9
)

var (
	SwapMode_name = map[int32]string{
		0: "disabled",
		1: "by_points",
		2: "by_symbol_currency",
		3: "by_margin_currency",
		4: "by_group_currency",
		5: "by_interest_current",
		6: "by_interest_open",
		7: "reopen_by_close_price",
		8: "reopen_by_bid",
		9: "by_profit_currency",
	}
	SwapMode_value = map[string]int32{
		"disabled":              0,
		"by_points":             1,
		"by_symbol_currency":    2,
		"by_margin_currency":    3,
		"by_group_currency":     4,
		"by_interest_current":   5,
		"by_interest_open":      6,
		"reopen_by_close_price": 7,
		"reopen_by_bid":         8,
		"by_profit_currency":    9,
	}
)

type SwapDays int32

const (
	SwapDays_sunday    SwapDays = 0
	SwapDays_monday    SwapDays = 1
	SwapDays_tuesday   SwapDays = 2
	SwapDays_wednesday SwapDays = 3
	SwapDays_thursday  SwapDays = 4
	SwapDays_friday    SwapDays = 5
	SwapDays_saturday  SwapDays = 6
	SwapDays_disabled  SwapDays = 7
)

var (
	SwapDays_name = map[int32]string{
		0: "sunday",
		1: "monday",
		2: "tuesday",
		3: "wednesday",
		4: "thursday",
		5: "friday",
		6: "saturday",
		7: "disabled",
	}
	SwapDays_value = map[string]int32{
		"sunday":    0,
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		"disabled":  7,
	}
)

type SwapFlags int32

const (
	SwapFlags_none              SwapFlags = 0
	SwapFlags_consider_holidays SwapFlags = 1
)

var (
	SwapFlags_name = map[int32]string{
		0: "none",
		1: "consider_holidays",
	}
	SwapFlags_value = map[string]int32{
		"none":              0,
		"consider_holidays": 1,
	}
)

// SymbolMarginFlags is the symbol-level EnMarginFlags (not group-level).
type SymbolMarginFlags int32

const (
	SymbolMarginFlags_none            SymbolMarginFlags = 0
	SymbolMarginFlags_check_process   SymbolMarginFlags = 1
	SymbolMarginFlags_check_sltp      SymbolMarginFlags = 2
	SymbolMarginFlags_hedge_large_leg SymbolMarginFlags = 4
	SymbolMarginFlags_exclude_pl      SymbolMarginFlags = 8
	SymbolMarginFlags_recalc_rates    SymbolMarginFlags = 16

	SymbolMarginFlags_all SymbolMarginFlags = 31
)

var (
	SymbolMarginFlags_name = map[int32]string{
		0:  "none",
		1:  "check_process",
		2:  "check_sltp",
		4:  "hedge_large_leg",
		8:  "exclude_pl",
		16: "recalc_rates",
		31: "all",
	}
	SymbolMarginFlags_value = map[string]int32{
		"none":            0,
		"check_process":   1,
		"check_sltp":      2,
		"hedge_large_leg": 4,
		"exclude_pl":      8,
		"recalc_rates":    16,
		"all":             31,
	}
)

type TickFlags int32

const (
	TickFlags_none       TickFlags = 0
	TickFlags_realtime   TickFlags = 1
	TickFlags_collectraw TickFlags = 2
	TickFlags_feed_stats TickFlags = 4
)

var (
	TickFlags_name = map[int32]string{
		0: "none",
		1: "realtime",
		2: "collectraw",
		4: "feed_stats",
		8: "negative_prices",
	}
	TickFlags_value = map[string]int32{
		"none":            0,
		"realtime":        1,
		"collectraw":      2,
		"feed_stats":      4,
		"negative_prices": 8,
	}
)

type ChartMode int32

const (
	ChartMode_bid_price  ChartMode = 0
	ChartMode_last_price ChartMode = 1
	ChartMode_old        ChartMode = 255
)

var (
	ChartMode_name = map[int32]string{
		0:   "bid_price",
		1:   "last_price",
		255: "old",
	}
	ChartMode_value = map[string]int32{
		"bid_price":  0,
		"last_price": 1,
		"old":        255,
	}
)

type OptionMode int32

const (
	OptionMode_european_call OptionMode = 0
	OptionMode_european_put  OptionMode = 1
	OptionMode_american_call OptionMode = 2
	OptionMode_american_put  OptionMode = 3
)

var (
	OptionMode_name = map[int32]string{
		0: "european_call",
		1: "european_put",
		2: "american_call",
		3: "american_put",
	}
	OptionMode_value = map[string]int32{
		"european_call": 0,
		"european_put":  1,
		"american_call": 2,
		"american_put":  3,
	}
)

type SpliceType int32

const (
	SpliceType_none       SpliceType = 0
	SpliceType_unadjusted SpliceType = 1
	SpliceType_adjusted   SpliceType = 2
)

var (
	SpliceType_name = map[int32]string{
		0: "none",
		1: "unadjusted",
		2: "adjusted",
	}
	SpliceType_value = map[string]int32{
		"none":       0,
		"unadjusted": 1,
		"adjusted":   2,
	}
)

type SpliceTimeType int32

const (
	SpliceTimeType_expiration SpliceTimeType = 0
)

var (
	SpliceTimeType_name = map[int32]string{
		0: "expiration",
	}
	SpliceTimeType_value = map[string]int32{
		"expiration": 0,
	}
)

type InstantMode int32

const (
	InstantMode_check_normal InstantMode = 0
)

var (
	InstantMode_name = map[int32]string{
		0: "check_normal",
	}
	InstantMode_value = map[string]int32{
		"check_normal": 0,
	}
)

// InstantFlags is EnInstantFlags: what the instant execution mode may do.
type InstantFlags int32

const (
	InstantFlags_none              InstantFlags = 0
	InstantFlags_fast_confirmation InstantFlags = 1
)

var (
	InstantFlags_name = map[int32]string{
		0: "none",
		1: "fast_confirmation",
	}
	InstantFlags_value = map[string]int32{
		"none":              0,
		"fast_confirmation": 1,
	}
)

type RequestFlags int32

const (
	RequestFlags_none  RequestFlags = 0
	RequestFlags_order RequestFlags = 1
)

var (
	RequestFlags_name = map[int32]string{
		0: "none",
		1: "order",
	}
	RequestFlags_value = map[string]int32{
		"none":  0,
		"order": 1,
	}
)

// SymbolTradeFlags is the symbol-level EnTradeFlags (not group-level).
type SymbolTradeFlags int32

const (
	SymbolTradeFlags_none             SymbolTradeFlags = 0
	SymbolTradeFlags_profit_by_market SymbolTradeFlags = 1
	SymbolTradeFlags_allow_signals    SymbolTradeFlags = 2
)

var (
	SymbolTradeFlags_name = map[int32]string{
		0: "none",
		1: "profit_by_market",
		2: "allow_signals",
	}
	SymbolTradeFlags_value = map[string]int32{
		"none":             0,
		"profit_by_market": 1,
		"allow_signals":    2,
	}
)

type SymbolSector int32

const (
	SymbolSector_undefined              SymbolSector = 0
	SymbolSector_basic_materials        SymbolSector = 1
	SymbolSector_communication_services SymbolSector = 2
	SymbolSector_consumer_cyclical      SymbolSector = 3
	SymbolSector_consumer_defensive     SymbolSector = 4
	SymbolSector_energy                 SymbolSector = 5
	SymbolSector_financial              SymbolSector = 6
	SymbolSector_healthcare             SymbolSector = 7
	SymbolSector_industrials            SymbolSector = 8
	SymbolSector_real_estate            SymbolSector = 9
	SymbolSector_technology             SymbolSector = 10
	SymbolSector_utilities              SymbolSector = 11
	SymbolSector_currency               SymbolSector = 12
	SymbolSector_currency_crypto        SymbolSector = 13
	SymbolSector_indexes                SymbolSector = 14
	SymbolSector_commodities            SymbolSector = 15
)

var (
	SymbolSector_name = map[int32]string{
		0:  "undefined",
		1:  "basic_materials",
		2:  "communication_services",
		3:  "consumer_cyclical",
		4:  "consumer_defensive",
		5:  "energy",
		6:  "financial",
		7:  "healthcare",
		8:  "industrials",
		9:  "real_estate",
		10: "technology",
		11: "utilities",
		12: "currency",
		13: "currency_crypto",
		14: "indexes",
		15: "commodities",
	}
	SymbolSector_value = map[string]int32{
		"undefined":              0,
		"basic_materials":        1,
		"communication_services": 2,
		"consumer_cyclical":      3,
		"consumer_defensive":     4,
		"energy":                 5,
		"financial":              6,
		"healthcare":             7,
		"industrials":            8,
		"real_estate":            9,
		"technology":             10,
		"utilities":              11,
		"currency":               12,
		"currency_crypto":        13,
		"indexes":                14,
		"commodities":            15,
	}
)

type SymbolSessionType int32

const (
	SymbolSessionType_quote SymbolSessionType = 0
	SymbolSessionType_trade SymbolSessionType = 1
)

var (
	SymbolSessionType_name = map[int32]string{
		0: "quote",
		1: "trade",
	}
	SymbolSessionType_value = map[string]int32{
		"quote": 0,
		"trade": 1,
	}
)

// Symbol is the server-wide instrument master.
type Symbol struct {
	SymbolId             int64          `db:"symbol_id" json:"symbol_id"`
	Symbol               string         `db:"symbol" json:"symbol"`
	Path                 string         `db:"path" json:"path"`
	Isin                 string         `db:"isin" json:"isin"`
	Description          string         `db:"description" json:"description"`
	International        string         `db:"international" json:"international"`
	Category             string         `db:"category" json:"category"`
	Exchange             string         `db:"exchange" json:"exchange"`
	Cfi                  string         `db:"cfi" json:"cfi"`
	Sector               SymbolSector   `db:"sector" json:"sector"`
	Industry             SymbolIndustry `db:"industry" json:"industry"`
	Country              string         `db:"country" json:"country"`
	Basis                string         `db:"basis" json:"basis"`
	Source               string         `db:"source" json:"source"`
	Page                 string         `db:"page" json:"page"`
	CurrencyBase         string         `db:"currency_base" json:"currency_base"`
	CurrencyBaseDigits   int32          `db:"currency_base_digits" json:"currency_base_digits"`
	CurrencyProfit       string         `db:"currency_profit" json:"currency_profit"`
	CurrencyProfitDigits int32          `db:"currency_profit_digits" json:"currency_profit_digits"`
	CurrencyMargin       string         `db:"currency_margin" json:"currency_margin"`
	CurrencyMarginDigits int32          `db:"currency_margin_digits" json:"currency_margin_digits"`
	Color                int64          `db:"color" json:"color"`
	ColorBackground      int64          `db:"color_background" json:"color_background"`
	Digits               int32          `db:"digits" json:"digits"`
	Point                float64        `db:"point" json:"point"`
	Multiply             float64        `db:"multiply" json:"multiply"`
	TickFlags            TickFlags      `db:"tick_flags" json:"tick_flags"`
	// TickBookDepth > 0: exchange DOM; Spread/SpreadBalance are not applied (see hst-quote ApplySpread).
	TickBookDepth                  int32             `db:"tick_book_depth" json:"tick_book_depth"`
	TickBookVolume                 int32             `db:"tick_book_volume" json:"tick_book_volume"`
	FilterSoft                     int32             `db:"filter_soft" json:"filter_soft"`
	FilterSoftTicks                int32             `db:"filter_soft_ticks" json:"filter_soft_ticks"`
	FilterHard                     int32             `db:"filter_hard" json:"filter_hard"`
	FilterHardTicks                int32             `db:"filter_hard_ticks" json:"filter_hard_ticks"`
	FilterDiscard                  int32             `db:"filter_discard" json:"filter_discard"`
	FilterSpreadMax                int32             `db:"filter_spread_max" json:"filter_spread_max"`
	FilterSpreadMin                int32             `db:"filter_spread_min" json:"filter_spread_min"`
	SubscriptionsDelay             int32             `db:"subscriptions_delay" json:"subscriptions_delay"`
	TradeMode                      TradeMode         `db:"trade_mode" json:"trade_mode"`
	CalcMode                       CalcMode          `db:"calc_mode" json:"calc_mode"`
	ExecMode                       ExecMode          `db:"exec_mode" json:"exec_mode"`
	GtcMode                        GTCMode           `db:"gtc_mode" json:"gtc_mode"`
	FillFlags                      FillingFlags      `db:"fill_flags" json:"fill_flags"`
	ExpirFlags                     ExpirationFlags   `db:"expir_flags" json:"expir_flags"`
	Spread                         int32             `db:"spread" json:"spread"`
	SpreadBalance                  int32             `db:"spread_balance" json:"spread_balance"`
	SpreadDiff                     int32             `db:"spread_diff" json:"spread_diff"`
	SpreadDiffBalance              int32             `db:"spread_diff_balance" json:"spread_diff_balance"`
	TickValue                      float64           `db:"tick_value" json:"tick_value"`
	TickSize                       float64           `db:"tick_size" json:"tick_size"`
	ContractSize                   float64           `db:"contract_size" json:"contract_size"`
	StopsLevel                     int32             `db:"stops_level" json:"stops_level"`
	FreezeLevel                    int32             `db:"freeze_level" json:"freeze_level"`
	QuotesTimeout                  int32             `db:"quotes_timeout" json:"quotes_timeout"`
	VolumeMin                      int64             `db:"volume_min" json:"volume_min"`
	VolumeMinExt                   int64             `db:"volume_min_ext" json:"volume_min_ext"`
	VolumeMax                      int64             `db:"volume_max" json:"volume_max"`
	VolumeMaxExt                   int64             `db:"volume_max_ext" json:"volume_max_ext"`
	VolumeStep                     int64             `db:"volume_step" json:"volume_step"`
	VolumeStepExt                  int64             `db:"volume_step_ext" json:"volume_step_ext"`
	VolumeLimit                    int64             `db:"volume_limit" json:"volume_limit"`
	VolumeLimitExt                 int64             `db:"volume_limit_ext" json:"volume_limit_ext"`
	MarginFlags                    SymbolMarginFlags `db:"margin_flags" json:"margin_flags"`
	MarginInitial                  float64           `db:"margin_initial" json:"margin_initial"`
	MarginMaintenance              float64           `db:"margin_maintenance" json:"margin_maintenance"`
	MarginInitialBuy               float64           `db:"margin_initial_buy" json:"margin_initial_buy"`
	MarginInitialSell              float64           `db:"margin_initial_sell" json:"margin_initial_sell"`
	MarginInitialBuyLimit          float64           `db:"margin_initial_buy_limit" json:"margin_initial_buy_limit"`
	MarginInitialSellLimit         float64           `db:"margin_initial_sell_limit" json:"margin_initial_sell_limit"`
	MarginInitialBuyStop           float64           `db:"margin_initial_buy_stop" json:"margin_initial_buy_stop"`
	MarginInitialSellStop          float64           `db:"margin_initial_sell_stop" json:"margin_initial_sell_stop"`
	MarginInitialBuyStopLimit      float64           `db:"margin_initial_buy_stop_limit" json:"margin_initial_buy_stop_limit"`
	MarginInitialSellStopLimit     float64           `db:"margin_initial_sell_stop_limit" json:"margin_initial_sell_stop_limit"`
	MarginMaintenanceBuy           float64           `db:"margin_maintenance_buy" json:"margin_maintenance_buy"`
	MarginMaintenanceSell          float64           `db:"margin_maintenance_sell" json:"margin_maintenance_sell"`
	MarginMaintenanceBuyLimit      float64           `db:"margin_maintenance_buy_limit" json:"margin_maintenance_buy_limit"`
	MarginMaintenanceSellLimit     float64           `db:"margin_maintenance_sell_limit" json:"margin_maintenance_sell_limit"`
	MarginMaintenanceBuyStop       float64           `db:"margin_maintenance_buy_stop" json:"margin_maintenance_buy_stop"`
	MarginMaintenanceSellStop      float64           `db:"margin_maintenance_sell_stop" json:"margin_maintenance_sell_stop"`
	MarginMaintenanceBuyStopLimit  float64           `db:"margin_maintenance_buy_stop_limit" json:"margin_maintenance_buy_stop_limit"`
	MarginMaintenanceSellStopLimit float64           `db:"margin_maintenance_sell_stop_limit" json:"margin_maintenance_sell_stop_limit"`
	MarginHedged                   float64           `db:"margin_hedged" json:"margin_hedged"`
	SwapMode                       SwapMode          `db:"swap_mode" json:"swap_mode"`
	SwapLong                       float64           `db:"swap_long" json:"swap_long"`
	SwapShort                      float64           `db:"swap_short" json:"swap_short"`
	SwapYearDay                    int32             `db:"swap_year_day" json:"swap_year_day"`
	SwapFlags                      SwapFlags         `db:"swap_flags" json:"swap_flags"`
	SwapRateSunday                 float64           `db:"swap_rate_sunday" json:"swap_rate_sunday"`
	SwapRateMonday                 float64           `db:"swap_rate_monday" json:"swap_rate_monday"`
	SwapRateTuesday                float64           `db:"swap_rate_tuesday" json:"swap_rate_tuesday"`
	SwapRateWednesday              float64           `db:"swap_rate_wednesday" json:"swap_rate_wednesday"`
	SwapRateThursday               float64           `db:"swap_rate_thursday" json:"swap_rate_thursday"`
	SwapRateFriday                 float64           `db:"swap_rate_friday" json:"swap_rate_friday"`
	SwapRateSaturday               float64           `db:"swap_rate_saturday" json:"swap_rate_saturday"`
	TimeStart                      int64             `db:"time_start" json:"time_start"`
	TimeExpiration                 int64             `db:"time_expiration" json:"time_expiration"`
	ReFlags                        RequestFlags      `db:"re_flags" json:"re_flags"`
	ReTimeout                      int32             `db:"re_timeout" json:"re_timeout"`
	IeCheckMode                    InstantMode       `db:"ie_check_mode" json:"ie_check_mode"`
	IeTimeout                      int32             `db:"ie_timeout" json:"ie_timeout"`
	IeSlipProfit                   int32             `db:"ie_slip_profit" json:"ie_slip_profit"`
	IeFlags                        InstantFlags      `db:"ie_flags" json:"ie_flags"`
	IeSlipLosing                   int32             `db:"ie_slip_losing" json:"ie_slip_losing"`
	IeVolumeMax                    int64             `db:"ie_volume_max" json:"ie_volume_max"`
	IeVolumeMaxExt                 int64             `db:"ie_volume_max_ext" json:"ie_volume_max_ext"`
	PriceSettle                    float64           `db:"price_settle" json:"price_settle"`
	PriceLimitMax                  float64           `db:"price_limit_max" json:"price_limit_max"`
	PriceLimitMin                  float64           `db:"price_limit_min" json:"price_limit_min"`
	TradeFlags                     SymbolTradeFlags  `db:"trade_flags" json:"trade_flags"`
	OrderFlags                     OrderFlags        `db:"order_flags" json:"order_flags"`
	MarginRateLiquidity            float64           `db:"margin_rate_liquidity" json:"margin_rate_liquidity"`
	MarginRateCurrency             float64           `db:"margin_rate_currency" json:"margin_rate_currency"`
	FaceValue                      float64           `db:"face_value" json:"face_value"`
	AccruedInterest                float64           `db:"accrued_interest" json:"accrued_interest"`
	SpliceType                     SpliceType        `db:"splice_type" json:"splice_type"`
	SpliceTimeType                 SpliceTimeType    `db:"splice_time_type" json:"splice_time_type"`
	SpliceTimeDays                 int32             `db:"splice_time_days" json:"splice_time_days"`
	OptionMode                     OptionMode        `db:"option_mode" json:"option_mode"`
	PriceStrike                    float64           `db:"price_strike" json:"price_strike"`
	FilterGap                      int32             `db:"filter_gap" json:"filter_gap"`
	FilterGapTicks                 int32             `db:"filter_gap_ticks" json:"filter_gap_ticks"`
	TickChartMode                  ChartMode         `db:"tick_chart_mode" json:"tick_chart_mode"`
	DateCreated                    int64             `db:"date_created" json:"date_created"`
	DateModified                   int64             `db:"date_modified" json:"date_modified"`
}

func (Symbol) TableName() string { return "hst.symbols" }

// SymbolSession is one quote or trade window for a symbol weekday.
type SymbolSession struct {
	SessionId int64             `db:"session_id" json:"session_id"`
	SymbolId  int64             `db:"symbol_id" json:"symbol_id"`
	Type      SymbolSessionType `db:"type" json:"type"`
	Day       int16             `db:"day" json:"day"`
	Open      int32             `db:"open" json:"open"`
	Close     int32             `db:"close" json:"close"`
}

func (SymbolSession) TableName() string { return "hst.symbols_sessions" }

// How the margin for an instrument is worked out.
type CalcMode int32

const (
	CalcMode_forex               CalcMode = 0
	CalcMode_futures             CalcMode = 1
	CalcMode_cfd                 CalcMode = 2
	CalcMode_cfd_index           CalcMode = 3
	CalcMode_cfd_leverage        CalcMode = 4
	CalcMode_forex_no_leverage   CalcMode = 5
	CalcMode_exch_stocks         CalcMode = 32
	CalcMode_exch_futures        CalcMode = 33
	CalcMode_exch_forts          CalcMode = 34
	CalcMode_exch_options        CalcMode = 35
	CalcMode_exch_options_margin CalcMode = 36
	CalcMode_exch_bonds          CalcMode = 37
	CalcMode_serv_collateral     CalcMode = 64
)

// How a request is turned into a fill.
type ExecMode int32

const (
	ExecMode_request  ExecMode = 0
	ExecMode_instant  ExecMode = 1
	ExecMode_market   ExecMode = 2
	ExecMode_exchange ExecMode = 3
)

// What a group may do with an instrument.
type TradeMode int32

const (
	TradeMode_disabled   TradeMode = 0
	TradeMode_long_only  TradeMode = 1
	TradeMode_short_only TradeMode = 2
	TradeMode_close_only TradeMode = 3
	TradeMode_full       TradeMode = 4
)
