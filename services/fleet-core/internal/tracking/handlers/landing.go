package handlers

import (
	"encoding/json"
	"net/http"
)

type SuccessStory struct {
	Icon       string `json:"icon"`
	Color      string `json:"color"`
	Value      string `json:"value"`
	Title      string `json:"title"`
	Quote      string `json:"quote"`
	Company    string `json:"company"`
	Fleet      string `json:"fleet"`
	Badge      string `json:"badge"`
	BadgeColor string `json:"badgeColor"`
}

type Feature struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Color string `json:"color"`
}

type Integration struct {
	Icon  string `json:"icon"`
	Label string `json:"label"`
	Color string `json:"color"`
}

type SecurityItem struct {
	Icon  string `json:"icon"`
	Label string `json:"label"`
	Desc  string `json:"desc"`
	Color string `json:"color"`
}

type HardwareAddon struct {
	Icon  string `json:"icon"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

type FAQ struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type TrustedBy struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type TimelineItem struct {
	Week  int    `json:"week"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type PricingPlan struct {
	Name      string   `json:"name"`
	Monthly   any      `json:"monthly"` // Can be int or string "Custom"
	Yearly    any      `json:"yearly"`  // Can be int or string "Custom"
	Features  []string `json:"features"`
	IsPopular bool     `json:"isPopular,omitempty"`
}

type LandingDataResponse struct {
	SuccessStories []SuccessStory  `json:"successStories"`
	Features       []Feature       `json:"features"`
	Integrations   []Integration   `json:"integrations"`
	SecurityItems  []SecurityItem  `json:"securityItems"`
	HardwareAddons []HardwareAddon `json:"hardwareAddons"`
	FAQs           []FAQ           `json:"faqs"`
	TrustedBy      []TrustedBy     `json:"trustedBy"`
	Timeline       []TimelineItem  `json:"timeline"`
	Pricing        []PricingPlan   `json:"pricing"`
}

type LandingHandler struct{}

func NewLandingHandler() *LandingHandler {
	return &LandingHandler{}
}

func (h *LandingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := LandingDataResponse{
		SuccessStories: []SuccessStory{
			{
				Icon:       "trending_down",
				Color:      "blue",
				Value:      "24%",
				Title:      "Reduced Fuel Costs",
				Quote:      "Global Logistics implemented our fuel monitoring solution across 450 trucks, identifying theft and idling inefficiencies within week one.",
				Company:    "Global Logistics Inc.",
				Fleet:      "Fleet of 450+ Vehicles",
				Badge:      "GL",
				BadgeColor: "blue",
			},
			{
				Icon:       "security",
				Color:      "green",
				Value:      "40%",
				Title:      "Fewer Accidents",
				Quote:      "By utilizing the Driver Scorecard module, CityTrans gamified safety, resulting in a massive drop in harsh braking events and insurance claims.",
				Company:    "CityTrans Delivery",
				Fleet:      "Fleet of 120+ Vans",
				Badge:      "CT",
				BadgeColor: "green",
			},
			{
				Icon:       "schedule",
				Color:      "orange",
				Value:      "15hr",
				Title:      "Weekly Admin Saved",
				Quote:      "Automating trip classification and expense reports saved the finance team 15 hours per week on manual data entry.",
				Company:    "BuildCorp Construction",
				Fleet:      "Heavy Equipment Fleet",
				Badge:      "BC",
				BadgeColor: "orange",
			},
		},
		Features: []Feature{
			{
				Icon:  "location_on",
				Title: "Live Tracking",
				Desc:  "Real-time GPS coordinates with 5-second refresh rates and traffic overlays.",
				Color: "blue",
			},
			{
				Icon:  "route",
				Title: "Trip Management",
				Desc:  "Historical route playback, trip classification (business/personal), and stop duration logs.",
				Color: "purple",
			},
			{
				Icon:  "share_location",
				Title: "Smart Geofencing",
				Desc:  "Create unlimited custom zones and get instant alerts when vehicles enter or exit specific areas.",
				Color: "indigo",
			},
			{
				Icon:  "water_drop",
				Title: "Fuel Monitoring",
				Desc:  "Track consumption patterns, detect theft events, and optimize refueling stops.",
				Color: "red",
			},
			{
				Icon:  "car_repair",
				Title: "Maintenance Alerts",
				Desc:  "Automated reminders for oil changes, tire rotations, and insurance renewals.",
				Color: "orange",
			},
			{
				Icon:  "speed",
				Title: "Driver Scorecards",
				Desc:  "Gamify safety with driver scores based on speeding, harsh braking, and cornering.",
				Color: "green",
			},
		},
		Integrations: []Integration{
			{Icon: "database", Label: "SAP / ERP", Color: "text-primary"},
			{Icon: "chat", Label: "WhatsApp API", Color: "text-green-500"},
			{Icon: "local_gas_station", Label: "Fuel Cards", Color: "text-blue-500"},
			{Icon: "toll", Label: "FastTag", Color: "text-purple-500"},
			{Icon: "map", Label: "Google Maps", Color: "text-orange-500"},
			{Icon: "webhook", Label: "Webhooks", Color: "text-indigo-500"},
		},
		SecurityItems: []SecurityItem{
			{
				Icon:  "verified_user",
				Label: "ISO 27001",
				Desc:  "Certified information security management systems.",
				Color: "blue",
			},
			{
				Icon:  "shield",
				Label: "GDPR Compliant",
				Desc:  "Full compliance with European data privacy regulations.",
				Color: "green",
			},
			{
				Icon:  "lock",
				Label: "SOC 2 Type II",
				Desc:  "Audited controls for security, availability, and confidentiality.",
				Color: "purple",
			},
			{
				Icon:  "timer",
				Label: "99.9% SLA",
				Desc:  "Guaranteed uptime with financial backing for enterprise plans.",
				Color: "orange",
			},
		},
		HardwareAddons: []HardwareAddon{
			{Icon: "videocam", Name: "AI Dashcam", Price: "$199"},
			{Icon: "thermostat", Name: "Temp Sensor", Price: "$49"},
			{Icon: "door_sensor", Name: "Door Sensor", Price: "$39"},
			{Icon: "lock", Name: "Remote Immobilizer", Price: "$149"},
		},
		FAQs: []FAQ{
			{
				Q: "How difficult is installation?",
				A: "Extremely simple. Our plug-and-play OBD-II devices install in seconds.",
			},
			{
				Q: "Is my data secure?",
				A: "We are ISO 27001 certified and GDPR compliant. All data is encrypted.",
			},
			{
				Q: "What if a driver unplugs the device?",
				A: "You will receive an instant 'Device Unplugged' alert via SMS and email.",
			},
			{
				Q: "What support options are available?",
				A: "All plans include email support. Pro gets chat, Enterprise gets 24/7 phone.",
			},
		},
		TrustedBy: []TrustedBy{
			{Name: "ATLANTIC", Icon: "anchor"},
			{Name: "AIRCARGO", Icon: "flight_takeoff"},
			{Name: "TRANSROAD", Icon: "local_shipping"},
			{Name: "BOXLOGISTICS", Icon: "package_2"},
			{Name: "SUPPLYCHAIN", Icon: "warehouse"},
		},
		Timeline: []TimelineItem{
			{Week: 1, Title: "Analysis", Desc: "Hardware Shipment & Data Migration"},
			{Week: 2, Title: "Config", Desc: "Legacy Integration & Geofence Setup"},
			{Week: 3, Title: "Install", Desc: "Hardware Rollout & Testing"},
			{Week: 4, Title: "Go-Live", Desc: "Staff Training & Full Launch"},
		},
		Pricing: []PricingPlan{
			{
				Name:     "Starter",
				Monthly:  29,
				Yearly:   24,
				Features: []string{"Real-time GPS Tracking", "Basic Reports", "Email Support"},
			},
			{
				Name:      "Pro",
				Monthly:   49,
				Yearly:    39,
				Features:  []string{"Everything in Starter", "AI Dashcam Integration", "Priority Support"},
				IsPopular: true,
			},
			{
				Name:     "Enterprise",
				Monthly:  "Custom",
				Yearly:   "Custom",
				Features: []string{"Unlimited Vehicles", "Dedicated Account Manager", "24/7 Phone Support"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
