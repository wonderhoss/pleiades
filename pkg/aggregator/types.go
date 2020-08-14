package aggregator

// Server consumes events from filesystem or kafka, then calculates aggregate stats and stores them in redis
type Server interface {
	Start() error
	Stop()
}

// Meta comment
type Meta struct {
	Domain    string `json:"domain,omitempty"`
	DateTime  string `json:"dt"`
	ID        string `json:"id"`
	RequestID string `json:"request_id,omitempty"`
	Stream    string `json:"stream"`
	URI       string `json:"uri,omitempty"`
}

// MediawikiRecentchange comment
type MediawikiRecentchange struct {
	ID                   int64                  `json:"id,omitempty"`
	Meta                 *Meta                  `json:"meta"`
	Schema               string                 `json:"$schema"`
	AdditionalProperties map[string]interface{} `json:"-,omitempty"`
	Timestamp            int                    `json:"timestamp,omitempty"`
	Wiki                 string                 `json:"wiki,omitempty"`

	Bot     bool    `json:"bot,omitempty"`
	Comment string  `json:"comment,omitempty"`
	Length  *Length `json:"length,omitempty"`

	LogAction        string      `json:"log_action,omitempty"`
	LogActionComment interface{} `json:"log_action_comment,omitempty"`
	LogID            interface{} `json:"log_id,omitempty"`
	LogParams        interface{} `json:"log_params,omitempty"`
	LogType          interface{} `json:"log_type,omitempty"`

	Minor         bool      `json:"minor,omitempty"`
	Namespace     int       `json:"namespace,omitempty"`
	Parsedcomment string    `json:"parsedcomment,omitempty"`
	Patrolled     bool      `json:"patrolled,omitempty"`
	Revision      *Revision `json:"revision,omitempty"`

	ServerName       string `json:"server_name,omitempty"`
	ServerScriptPath string `json:"server_script_path,omitempty"`
	ServerURL        string `json:"server_url,omitempty"`

	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
	User  string `json:"user,omitempty"`
}

// Length comment
type Length struct {
	New int64 `json:"new,omitempty"`
	Old int64 `json:"old,omitempty"`
}

// Revision comment
type Revision struct {
	New interface{} `json:"new,omitempty"`
	Old interface{} `json:"old,omitempty"`
}
