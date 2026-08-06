package model

// SymbolIndustry stores MT5 ENUM_SYMBOL_INDUSTRY wire values.
type SymbolIndustry int32

const (
	SymbolIndustry_undefined                    SymbolIndustry = 0
	SymbolIndustry_agricultural_inputs                         = 1
	SymbolIndustry_aluminium                                   = 2
	SymbolIndustry_building_materials                          = 3
	SymbolIndustry_chemicals                                   = 4
	SymbolIndustry_coking_coal                                 = 5
	SymbolIndustry_copper                                      = 6
	SymbolIndustry_gold                                        = 7
	SymbolIndustry_lumber_wood                                 = 8
	SymbolIndustry_industrial_metals                           = 9
	SymbolIndustry_precious_metals                             = 10
	SymbolIndustry_paper                                       = 11
	SymbolIndustry_silver                                      = 12
	SymbolIndustry_specialty_chemicals                         = 13
	SymbolIndustry_steel                                       = 14
	SymbolIndustry_advertising                                 = 15
	SymbolIndustry_broadcasting                                = 16
	SymbolIndustry_gaming_multimedia                           = 17
	SymbolIndustry_entertainment                               = 18
	SymbolIndustry_internet_content                            = 19
	SymbolIndustry_publishing                                  = 20
	SymbolIndustry_telecom                                     = 21
	SymbolIndustry_apparel_manufacturing                       = 22
	SymbolIndustry_apparel_retail                              = 23
	SymbolIndustry_auto_manufacturers                          = 24
	SymbolIndustry_auto_parts                                  = 25
	SymbolIndustry_auto_dealership                             = 26
	SymbolIndustry_department_stores                           = 27
	SymbolIndustry_footwear_accessories                        = 28
	SymbolIndustry_furnishings                                 = 29
	SymbolIndustry_gambling                                    = 30
	SymbolIndustry_home_improv_retail                          = 31
	SymbolIndustry_internet_retail                             = 32
	SymbolIndustry_leisure                                     = 33
	SymbolIndustry_lodging                                     = 34
	SymbolIndustry_luxury_goods                                = 35
	SymbolIndustry_packaging_containers                        = 36
	SymbolIndustry_personal_services                           = 37
	SymbolIndustry_recreational_vehicles                       = 38
	SymbolIndustry_resident_construction                       = 39
	SymbolIndustry_resorts_casinos                             = 40
	SymbolIndustry_restaurants                                 = 41
	SymbolIndustry_specialty_retail                            = 42
	SymbolIndustry_textile_manufacturing                       = 43
	SymbolIndustry_travel_services                             = 44
	SymbolIndustry_beverages_brewers                           = 45
	SymbolIndustry_beverages_non_alco                          = 46
	SymbolIndustry_beverages_wineries                          = 47
	SymbolIndustry_confectioners                               = 48
	SymbolIndustry_discount_stores                             = 49
	SymbolIndustry_education_trainig                           = 50
	SymbolIndustry_farm_products                               = 51
	SymbolIndustry_food_distribution                           = 52
	SymbolIndustry_grocery_stores                              = 53
	SymbolIndustry_household_products                          = 54
	SymbolIndustry_packaged_foods                              = 55
	SymbolIndustry_tobacco                                     = 56
	SymbolIndustry_oil_gas_drilling                            = 57
	SymbolIndustry_oil_gas_ep                                  = 58
	SymbolIndustry_oil_gas_equipment                           = 59
	SymbolIndustry_oil_gas_integrated                          = 60
	SymbolIndustry_oil_gas_midstream                           = 61
	SymbolIndustry_oil_gas_refining                            = 62
	SymbolIndustry_thermal_coal                                = 63
	SymbolIndustry_uranium                                     = 64
	SymbolIndustry_exchange_traded_fund                        = 65
	SymbolIndustry_assets_management                           = 66
	SymbolIndustry_banks_diversified                           = 67
	SymbolIndustry_banks_regional                              = 68
	SymbolIndustry_capital_markets                             = 69
	SymbolIndustry_close_end_fund_debt                         = 70
	SymbolIndustry_close_end_fund_equity                       = 71
	SymbolIndustry_close_end_fund_foreign                      = 72
	SymbolIndustry_credit_services                             = 73
	SymbolIndustry_financial_conglomerate                      = 74
	SymbolIndustry_financial_data_exchange                     = 75
	SymbolIndustry_insurance_brokers                           = 76
	SymbolIndustry_insurance_diversified                       = 77
	SymbolIndustry_insurance_life                              = 78
	SymbolIndustry_insurance_property                          = 79
	SymbolIndustry_insurance_reinsurance                       = 80
	SymbolIndustry_insurance_specialty                         = 81
	SymbolIndustry_mortgage_finance                            = 82
	SymbolIndustry_shell_companies                             = 83
	SymbolIndustry_biotechnology                               = 84
	SymbolIndustry_diagnostics_research                        = 85
	SymbolIndustry_drugs_manufacturers                         = 86
	SymbolIndustry_drugs_manufacturers_spec                    = 87
	SymbolIndustry_healthcare_plans                            = 88
	SymbolIndustry_health_information                          = 89
	SymbolIndustry_medical_facilities                          = 90
	SymbolIndustry_medical_devices                             = 91
	SymbolIndustry_medical_distribution                        = 92
	SymbolIndustry_medical_instruments                         = 93
	SymbolIndustry_pharm_retailers                             = 94
	SymbolIndustry_aerospace_defense                           = 95
	SymbolIndustry_airlines                                    = 96
	SymbolIndustry_airports_services                           = 97
	SymbolIndustry_building_products                           = 98
	SymbolIndustry_business_equipment                          = 99
	SymbolIndustry_conglomerates                               = 100
	SymbolIndustry_consulting_services                         = 101
	SymbolIndustry_electrical_equipment                        = 102
	SymbolIndustry_engineering_construction                    = 103
	SymbolIndustry_farm_heavy_machinery                        = 104
	SymbolIndustry_industrial_distribution                     = 105
	SymbolIndustry_infrastructure_operations                   = 106
	SymbolIndustry_freight_logistics                           = 107
	SymbolIndustry_marine_shipping                             = 108
	SymbolIndustry_metal_fabrication                           = 109
	SymbolIndustry_pollution_control                           = 110
	SymbolIndustry_railroads                                   = 111
	SymbolIndustry_rental_leasing                              = 112
	SymbolIndustry_security_protection                         = 113
	SymbolIndustry_speality_business_services                  = 114
	SymbolIndustry_speality_machinery                          = 115
	SymbolIndustry_stuffing_employment                         = 116
	SymbolIndustry_tools_accessories                           = 117
	SymbolIndustry_trucking                                    = 118
	SymbolIndustry_waste_management                            = 119
	SymbolIndustry_real_estate_development                     = 120
	SymbolIndustry_real_estate_diversified                     = 121
	SymbolIndustry_real_estate_services                        = 122
	SymbolIndustry_reit_diversified                            = 123
	SymbolIndustry_reit_healtcare                              = 124
	SymbolIndustry_reit_hotel_motel                            = 125
	SymbolIndustry_reit_industrial                             = 126
	SymbolIndustry_reit_mortage                                = 127
	SymbolIndustry_reit_office                                 = 128
	SymbolIndustry_reit_residental                             = 129
	SymbolIndustry_reit_retail                                 = 130
	SymbolIndustry_reit_speciality                             = 131
	SymbolIndustry_communication_equipment                     = 132
	SymbolIndustry_computer_hardware                           = 133
	SymbolIndustry_consumer_electronics                        = 134
	SymbolIndustry_electronic_components                       = 135
	SymbolIndustry_electronic_distribution                     = 136
	SymbolIndustry_it_services                                 = 137
	SymbolIndustry_scientific_instruments                      = 138
	SymbolIndustry_semiconductor_equipment                     = 139
	SymbolIndustry_semiconductors                              = 140
	SymbolIndustry_software_application                        = 141
	SymbolIndustry_software_infrastructure                     = 142
	SymbolIndustry_solar                                       = 143
	SymbolIndustry_utilities_diversified                       = 144
	SymbolIndustry_utilities_powerproducers                    = 145
	SymbolIndustry_utilities_renewable                         = 146
	SymbolIndustry_utilities_regulated_electric                = 147
	SymbolIndustry_utilities_regulated_gas                     = 148
	SymbolIndustry_utilities_regulated_water                   = 149
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
