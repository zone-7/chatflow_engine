package provider

const (
	MESSAGE_ROLE_USER      = "user"
	MESSAGE_ROLE_ASSISTANT = "assistant"
	MESSAGE_ROLE_SYSTEM    = "system"
)

var chattings = []string{}
var embeddings = []string{}
var vectordbs = []string{}

type Dict struct {
	Name   string  `json:"name" yaml:"name"`
	Fields []Field `json:"fields" yaml:"fields"`
}

type FieldOption struct {
	Value string `json:"value" yaml:"value"`
	Label string `json:"label" yaml:"label"`
}

type Field struct {
	Name         string `json:"name" yaml:"name"`
	Label        string `json:"label" yaml:"label"`
	Value        string `json:"value" yaml:"value"`
	DefaultValue string `json:"default_value" yaml:"default_value"`

	InputType string        `json:"input_type" yaml:"input_type"`
	Options   []FieldOption `json:"options" yaml:"options"`
}

// Embedding

type Embedding interface {
	GetDict() Dict
	Embed(params map[string]string, contents []string) ([][]float64, error)
}

func GetEmbeddingDicts() []Dict {
	dicts := []Dict{}
	for _, name := range embeddings {
		e := CreateEmbedding(name)
		dicts = append(dicts, e.GetDict())
	}

	return dicts
}

// chatting
type ChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Partial bool     `json:"partial"`
}

type Chatting interface {
	GetDict() Dict
	Chat(params map[string]string, messages []ChatMessage, callback func(msg []ChatMessage, is_done bool) error, is_suspend func() bool) error
}

func GetChattingDicts() []Dict {
	dicts := []Dict{}
	for _, name := range chattings {
		e := CreateChatting(name)
		dicts = append(dicts, e.GetDict())
	}

	return dicts
}

// VectorDB
type VectorData struct {
	Id      string    `json:"id"`
	Payload any       `json:"payload"`
	Vector  []float64 `json:"vector"`
	Score   float64   `json:"score"`
}

type VectorDB interface {
	GetDict() Dict
	Search(params map[string]string, vector []float64, score float64, limit int) ([]*VectorData, error)
	Save(params map[string]string, datas []*VectorData) error
	Get(params map[string]string, id string) (*VectorData, error)
	Remove(params map[string]string, id string) error
	Clear(params map[string]string) error
}

func GetVectorDBDicts() []Dict {
	dicts := make([]Dict, 0)
	for _, name := range vectordbs {
		e := CreateVectorDB(name)
		dicts = append(dicts, e.GetDict())
	}

	return dicts
}

func CreateChatting(name string) Chatting {
	var chatting Chatting
	if name == "ollama" {
		chatting = &Chatting_ollama{}
	}
	if name == "openai" {
		chatting = &Chatting_openai{}
	}
	if name == "baidu" {
		chatting = &Chatting_baidu{}
	}
	if name == "kimi" {
		chatting = &Chatting_kimi{}
	}

	return chatting
}

func CreateEmbedding(name string) Embedding {
	var embedding Embedding
	if name == "ollama" {
		embedding = &Embedding_ollama{}
	}
	if name == "openai" {
		embedding = &Embedding_openai{}
	}

	return embedding
}

func CreateVectorDB(name string) VectorDB {
	var db VectorDB
	if name == "qdrant" {
		db = &VectorDB_qdrant{}
	}
	if name == "es8" {
		db = &VectorDB_es8{}
	}

	return db
}
