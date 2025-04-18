package design

// Color palette
var Colors = struct {
	Primary   string
	Secondary string
	Accent    string
	Success   string
	Warning   string
	Error     string

	Background struct {
		Default string
		Paper   string
		Dark    string
	}

	Text struct {
		Primary   string
		Secondary string
		Disabled  string
		Light     string
	}

	Gray struct {
		Gray50  string
		Gray100 string
		Gray200 string
		Gray300 string
		Gray400 string
		Gray500 string
		Gray600 string
		Gray700 string
		Gray800 string
		Gray900 string
	}
}{
	Primary:   "indigo-500",
	Secondary: "gray-700",
	Accent:    "indigo-400",
	Success:   "green-500",
	Warning:   "yellow-500",
	Error:     "red-500",

	Background: struct {
		Default string
		Paper   string
		Dark    string
	}{
		Default: "bg-[#191A21]",
		Paper:   "bg-gray-800",
		Dark:    "bg-gray-900",
	},

	Text: struct {
		Primary   string
		Secondary string
		Disabled  string
		Light     string
	}{
		Primary:   "text-white",
		Secondary: "text-gray-300",
		Disabled:  "text-gray-400",
		Light:     "text-gray-300",
	},

	Gray: struct {
		Gray50  string
		Gray100 string
		Gray200 string
		Gray300 string
		Gray400 string
		Gray500 string
		Gray600 string
		Gray700 string
		Gray800 string
		Gray900 string
	}{
		Gray50:  "gray-50",
		Gray100: "gray-100",
		Gray200: "gray-200",
		Gray300: "gray-300",
		Gray400: "gray-400",
		Gray500: "gray-500",
		Gray600: "gray-600",
		Gray700: "gray-700",
		Gray800: "gray-800",
		Gray900: "gray-900",
	},
}

// Typography
var Typography = struct {
	FontFamily string

	Heading struct {
		H1 string
		H2 string
		H3 string
		H4 string
	}

	Body struct {
		Large  string
		Medium string
		Small  string
	}
}{
	FontFamily: "font-sans",

	Heading: struct {
		H1 string
		H2 string
		H3 string
		H4 string
	}{
		H1: "text-3xl font-bold tracking-wide",
		H2: "text-2xl font-bold",
		H3: "text-xl font-semibold",
		H4: "text-lg font-medium",
	},

	Body: struct {
		Large  string
		Medium string
		Small  string
	}{
		Large:  "text-lg",
		Medium: "text-base",
		Small:  "text-sm",
	},
}

// Spacing
var Spacing = struct {
	XS  string
	S   string
	M   string
	L   string
	XL  string
	XXL string
}{
	XS:  "p-2",
	S:   "p-3",
	M:   "p-4",
	L:   "p-6",
	XL:  "p-8",
	XXL: "p-12",
}

// Border radius
var BorderRadius = struct {
	None   string
	Small  string
	Medium string
	Large  string
	Full   string
}{
	None:   "rounded-none",
	Small:  "rounded",
	Medium: "rounded-md",
	Large:  "rounded-lg",
	Full:   "rounded-full",
}

// Shadows
var Shadows = struct {
	None   string
	Small  string
	Medium string
	Large  string
}{
	None:   "shadow-none",
	Small:  "shadow",
	Medium: "shadow-md",
	Large:  "shadow-lg",
}
