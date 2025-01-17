package meta

const (
	CHUNK_TYPE_NONE   = "none"   //不切片
	CHUNK_TYPE_WINDOW = "window" //滑动窗口
	CHUNK_TYPE_SPLIT  = "split"  //分隔符

)

type KnowledgeInfo struct {
	Id              string            `json:"id" yaml:"id"`
	Title           string            `json:"title" yaml:"title"`
	Description     string            `json:"description"`
	FileCount       int               `json:"file_count" yaml:"file_count"`
	WordCount       int               `json:"word_count" yaml:"word_count"`
	PayloadCount    int               `json:"payload_count" yaml:"payload_count"`
	CreateTime      string            `json:"create_time" yaml:"create_time"`
	Score           float64           `json:"score" yaml:"score"`
	Limit           int               `json:"limit" yaml:"limit"`
	Index           int               `json:"index" yaml:"index"`
	Embedding       string            `json:"embedding" yaml:"embedding"`
	EmbeddingParams map[string]string `json:"embedding_params" yaml:"embedding_params"`
	Vectordb        string            `json:"vectordb" yaml:"vectordb"`
	VectordbParams  map[string]string `json:"vectordb_params" yaml:"vectordb_params"`

	ChunkTimeStart string `json:"chunk_time_start" yaml:"chunk_time_start"`
	ChunkTimeStop  string `json:"chunk_time_stop" yaml:"chunk_time_stop"`
	ChunkStatus    string `json:"chunk_status" yaml:"chunk_status"`
	ChunkProgress  string `json:"chunk_progress" yaml:"chunk_progress"`

	VectorTimeStart string `json:"vector_time_start" yaml:"vector_time_start"`
	VectorTimeStop  string `json:"vector_time_stop" yaml:"vector_time_stop"`
	VectorStatus    string `json:"vector_status" yaml:"vector_status"`
	VectorProgress  string `json:"vector_progress" yaml:"vector_progress"`

	StoreTimeStart string `json:"store_time_start" yaml:"store_time_start"`
	StoreTimeStop  string `json:"store_time_stop" yaml:"store_time_stop"`
	StoreStatus    string `json:"store_status" yaml:"store_status"`
	StoreProgress  string `json:"store_progress" yaml:"store_progress"`
}

type FileInfo struct {
	KnowledgeId  string            `json:"knowledge_id" yaml:"knowledge_id"`
	FileName     string            `json:"file_name" yaml:"file_name"`
	ChunkType    string            `json:"chunk_type" yaml:"chunk_type"`
	ChunkStep    int               `json:"chunk_step" yaml:"chunk_step"`
	ChunkSize    int               `json:"chunk_size" yaml:"chunk_size"`
	ChunkSplit   string            `json:"chunk_split" yaml:"chunk_split"`
	WordReplace  map[string]string `json:"word_replace" yaml:"word_replace"`
	FileExt      string            `json:"file_ext" yaml:"file_ext"`
	CreateTime   string            `json:"create_time" yaml:"create_time"`
	Size         int64             `json:"size" yaml:"size"`
	WordCount    int               `json:"word_count" yaml:"word_count"`
	PayloadCount int               `json:"payload_count" yaml:"payload_count"`
}

// 知识集分段数据
type Payload struct {
	Index       int    `json:"index" yaml:"index"`
	Text        string `json:"text" yaml:"text"`
	FileName    string `json:"file_name" yaml:"file_name"`
	KnowledgeId string `json:"knowledge_id" yaml:"knowledge_id"`
	CreateTime  string `json:"create_time" yaml:"create_time"`
}

// 向量数据
type Vector struct {
	Id          string    `json:"id" yaml:"id"`
	KnowledgeId string    `json:"knowledge_id" yaml:"knowledge_id"`
	FileName    string    `json:"file_name" yaml:"file_name"`
	Index       int       `json:"index" yaml:"index"`
	Text        string    `json:"text" yaml:"text"`
	CreateTime  string    `json:"create_time" yaml:"create_time"`
	Vector      []float64 `json:"vector" yaml:"vector"`
}

// 向量数据库Qdrant配置
type VDB_qdrant struct {
	Name          string `json:"name" yaml:"name"`
	VerctorSize   int    `json:"vector_size" yaml:"vector_size"`
	Address       string `json:"address" yaml:"address"`
	Port          string `json:"port" yaml:"port"`
	AutoCreate    bool   `json:"auto_create" yaml:"auto_create"`
	Collection    string `json:"collection" yaml:"collection"`
	DistanceModel string `json:"distance_model" yaml:"distance_model"`
}
