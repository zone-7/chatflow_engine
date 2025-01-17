package manager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/zone-7/chatflow_engine/engine/meta"
	"github.com/zone-7/chatflow_engine/engine/provider"
	"github.com/zone-7/chatflow_engine/engine/utils"
	"gopkg.in/yaml.v2"
)

type Task struct {
	KnowledgeId string
	Opt         meta.Option
}

func NewTask(opt meta.Option, knowledgeId string) Task {
	return Task{Opt: opt, KnowledgeId: knowledgeId}
}

var vector_chan chan Task = make(chan Task, 50)
var store_chan chan Task = make(chan Task, 50)

func init() {
	go vector_process()
	go store_process()
}

func vector_process() {
	fmt.Println("vector_process_start")
	for {

		task, ok := <-vector_chan
		if !ok {
			break
		}

		manager := KnowledgeManager{Opt: task.Opt}
		manager.GenerateVectors(task.KnowledgeId)
	}

	fmt.Println("vector_process_end")
}

func store_process() {
	fmt.Println("store_process_start")
	for {

		task, ok := <-store_chan
		if !ok {
			break
		}

		manager := NewKnowledgeManager(task.Opt)
		manager.StoreKnowledge(task.KnowledgeId)

	}

	fmt.Println("store_process_end")
}

func PostVectorKnowledge(task Task) {
	vector_chan <- task
}

func PostStoreKnowledge(task Task) {
	store_chan <- task
}

type KnowledgeManager struct {
	Opt meta.Option
}

func NewKnowledgeManager(opt meta.Option) KnowledgeManager {
	return KnowledgeManager{Opt: opt}
}

func (k *KnowledgeManager) GetKnowledgeDir() string {
	return GetKnowledgePath(k.Opt)
}

// 获取知识集合数量
func (k *KnowledgeManager) GetKnowledgeCount() int {
	dir := k.GetKnowledgeDir()

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return 0
	}
	entries, err := os.ReadDir(dir)
	return len(entries)
}

// 获取所有知识集合信息列表
func (k *KnowledgeManager) KnowledgeInfoQuery(title string) ([]*meta.KnowledgeInfo, error) {

	title = strings.Trim(title, " ")
	title = strings.ToLower(title)

	dir := k.GetKnowledgeDir()

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	list, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	knowledges := make([]*meta.KnowledgeInfo, 0)
	for _, item := range list {
		if !item.IsDir() {
			continue
		}

		infopath := path.Join(dir, item.Name(), "info.yaml")
		data, err := os.ReadFile(infopath)
		if err != nil {
			continue
		}

		var info meta.KnowledgeInfo
		err = yaml.Unmarshal(data, &info)
		if err != nil {
			continue
		}

		// if info.SysUserId != k.Opt.SysUserId {
		// 	continue
		// }

		if len(title) > 0 {
			if !(strings.Contains(strings.ToLower(info.Title), title) || strings.Contains(strings.ToLower(info.Description), title)) {
				continue
			}
		}

		knowledges = append(knowledges, &info)
	}

	sort.Slice(knowledges, func(i, j int) bool {
		return knowledges[i].Index < knowledges[j].Index
	})

	return knowledges, nil
}

// 获取所有知识集合信息列表
func (k *KnowledgeManager) KnowledgeInfoList() ([]*meta.KnowledgeInfo, error) {
	dir := k.GetKnowledgeDir()

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	list, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	knowledges := make([]*meta.KnowledgeInfo, 0)
	for _, item := range list {
		if !item.IsDir() {
			continue
		}

		infopath := path.Join(dir, item.Name(), "info.yaml")
		data, err := os.ReadFile(infopath)
		if err != nil {
			continue
		}
		var info meta.KnowledgeInfo
		err = yaml.Unmarshal(data, &info)
		if err != nil {
			continue
		}
		// if info.SysUserId != k.Opt.SysUserId {
		// 	continue
		// }

		knowledges = append(knowledges, &info)
	}

	sort.Slice(knowledges, func(i, j int) bool {
		return knowledges[i].Index < knowledges[j].Index
	})

	return knowledges, nil
}

// 获取集合信息
func (k *KnowledgeManager) GetKnowledgeInfo(knowledge_id string) (*meta.KnowledgeInfo, error) {
	dir := k.GetKnowledgeDir()

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	infopath := path.Join(dir, knowledge_id, "info.yaml")
	data, err := os.ReadFile(infopath)
	if err != nil {
		return nil, err
	}
	var info meta.KnowledgeInfo

	err = yaml.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// 设置知识集信息
func (k *KnowledgeManager) SetKnowledgeInfo(knowledgeInfo *meta.KnowledgeInfo) error {

	if len(knowledgeInfo.Id) == 0 {
		uid, _ := uuid.NewV4()
		id := strings.ReplaceAll(uid.String(), "-", "")
		knowledgeInfo.Id = id
		knowledgeInfo.CreateTime = time.Now().Format("2006-01-02 03:04:05")
	}
	// if len(knowledgeInfo.SysUserId) == 0 {
	// 	knowledgeInfo.SysUserId = k.Opt.SysUserId
	// }

	dir := path.Join(k.GetKnowledgeDir(), knowledgeInfo.Id)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(knowledgeInfo)
	if err != nil {
		return err
	}

	infopath := path.Join(dir, "info.yaml")
	err = os.WriteFile(infopath, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// 删除知识集合
func (k *KnowledgeManager) RemoveKnowledge(knowledge_id string) error {
	info, err := k.GetKnowledgeInfo(knowledge_id)
	if err != nil {
		return err
	}
	if info == nil {
		return errors.New("知识不存在")
	}
	filepath := path.Join(k.GetKnowledgeDir(), knowledge_id)
	err = os.RemoveAll(filepath)

	return err
}

// 获取集合切片数据
func (k *KnowledgeManager) GetKnowledgePayloads(knowledge_id string) ([]meta.Payload, error) {
	infos, err := k.GetFileInfos(knowledge_id)
	if err != nil {
		return nil, err
	}
	knowledge_payloads := make([]meta.Payload, 0)
	for _, info := range infos {
		payloads, err := k.GetFilePayloads(knowledge_id, info.FileName)
		if err != nil {
			continue
		}
		knowledge_payloads = append(knowledge_payloads, payloads...)
	}

	return knowledge_payloads, err
}

func (k *KnowledgeManager) GetKnowledgeVectors(knowledge_id string) ([]meta.Vector, error) {
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id, "vectors.yaml")

	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}

	var vectors []meta.Vector
	err = yaml.Unmarshal(data, &vectors)

	return vectors, err
}

// 转化为向量，保存到文件
func (k *KnowledgeManager) GenerateVectors(knowledge_id string) error {
	var err error

	knowledge, err := k.GetKnowledgeInfo(knowledge_id)
	if err != nil {
		return err
	}

	if knowledge.ChunkStatus != "OK" {
		return errors.New("知识库文件切片未完成，请先执行文件切片！")
	}
	if knowledge.VectorProgress == "BEGIN" {
		return errors.New("知识库正在执行向量化！")
	}

	knowledge.VectorTimeStart = time.Now().Format("2006-01-02 03:04:05")
	knowledge.VectorProgress = "BEGIN"
	knowledge.VectorStatus = ""
	k.SetKnowledgeInfo(knowledge)

	defer func() {
		knowledge.VectorProgress = "END"

		knowledge.VectorTimeStop = time.Now().Format("2006-01-02 03:04:05")
		if err != nil {
			knowledge.VectorStatus = err.Error()
		} else {
			knowledge.VectorStatus = "OK"
		}

		k.SetKnowledgeInfo(knowledge)
	}()

	if len(knowledge.Embedding) == 0 {
		return errors.New("向量模型未设置")
	}
	embedding := provider.CreateEmbedding(knowledge.Embedding)
	if embedding == nil {
		return errors.New("向量模型不存在")
	}

	payloads, err := k.GetKnowledgePayloads(knowledge_id)
	if err != nil {
		return err
	}

	var contents []string
	for _, payload := range payloads {
		contents = append(contents, payload.Text)
	}

	vectors, err := embedding.Embed(knowledge.EmbeddingParams, contents)
	if err != nil {
		return err
	}

	vs := make([]meta.Vector, 0)
	for index, vector := range vectors {
		payload := payloads[index]
		v := meta.Vector{}

		uid, _ := uuid.NewV4()

		v.Id = uid.String()
		v.KnowledgeId = payload.KnowledgeId
		v.FileName = payload.FileName
		v.Index = payload.Index
		v.Text = payload.Text
		v.Vector = vector

		vs = append(vs, v)
	}

	data, err := yaml.Marshal(vs)
	if err != nil {
		return err
	}

	//保存到文件
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	filepath := path.Join(dir, "vectors.yaml")
	err = os.WriteFile(filepath, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// 生成集合切片数据
func (k *KnowledgeManager) GeneratePayloads(knowledge_id string) (int, error) {
	var err error
	var payload_count int

	knowledge, err := k.GetKnowledgeInfo(knowledge_id)
	if err != nil {
		return 0, err
	}

	if knowledge.FileCount == 0 {
		return 0, errors.New("知识库没有文件，请先上传文件！")
	}
	if knowledge.ChunkProgress == "BEGIN" {
		return 0, errors.New("知识库正在切片！")
	}

	knowledge.ChunkTimeStart = time.Now().Format("2006-01-02 03:04:05")
	knowledge.ChunkProgress = "BEGIN"
	knowledge.ChunkStatus = ""
	k.SetKnowledgeInfo(knowledge)

	defer func() {
		knowledge.PayloadCount = payload_count
		knowledge.ChunkProgress = "END"
		knowledge.ChunkTimeStop = time.Now().Format("2006-01-02 03:04:05")
		if err != nil {
			knowledge.ChunkStatus = err.Error()
		} else {
			knowledge.ChunkStatus = "OK"
		}

		k.SetKnowledgeInfo(knowledge)

	}()

	files, err := k.GetFileInfos(knowledge_id)
	if err != nil {
		return 0, err
	}

	for index, file := range files {
		file_name := file.FileName
		file_payload_count, err := k.GenerateFilePayloads(knowledge_id, file_name)
		if err != nil {
			return 0, err
		}
		payload_count += file_payload_count

		knowledge.ChunkProgress = fmt.Sprintf("%v/%v", (index + 1), len(files))
	}

	return payload_count, nil
}

// 保存到向量库
func (k *KnowledgeManager) StoreKnowledge(knowledge_id string) error {
	var err error

	knowledge, err := k.GetKnowledgeInfo(knowledge_id)
	if err != nil {
		return err
	}

	if knowledge.VectorStatus != "OK" {
		return errors.New("知识库还未向量化，请先知识向量转化！")
	}
	if knowledge.StoreProgress == "BEGIN" {
		return errors.New("知识入库化正在执行！")
	}

	knowledge.StoreTimeStart = time.Now().Format("2006-01-02 03:04:05")
	knowledge.StoreProgress = "BEGIN"
	knowledge.StoreStatus = ""
	k.SetKnowledgeInfo(knowledge)

	defer func() {
		knowledge.StoreProgress = "END"
		knowledge.StoreTimeStop = time.Now().Format("2006-01-02 03:04:05")
		if err != nil {
			knowledge.StoreStatus = err.Error()
		} else {
			knowledge.StoreStatus = "OK"
		}

		k.SetKnowledgeInfo(knowledge)

	}()

	if len(knowledge.Vectordb) == 0 {
		return errors.New("向量数据库未设置")
	}
	vectordb := provider.CreateVectorDB(knowledge.Vectordb)
	if vectordb == nil {
		return errors.New("向量数据库不存在")
	}

	vectors, err := k.GetKnowledgeVectors(knowledge_id)
	if err != nil {
		return err
	}

	datas := make([]*provider.VectorData, 0)
	for _, vector := range vectors {

		data := &provider.VectorData{}

		data.Id = vector.Id
		payload := meta.Payload{KnowledgeId: vector.KnowledgeId, FileName: vector.FileName, Index: vector.Index, Text: vector.Text}

		data.Payload = payload

		data.Vector = vector.Vector
		datas = append(datas, data)
	}

	vectordb.Clear(knowledge.VectordbParams)

	err = vectordb.Save(knowledge.VectordbParams, datas)
	if err == nil {
		fmt.Println("导入数据库成功")
	} else {
		fmt.Println("导入数据库失败")
	}

	return err
}

// 从向量库中检索
func (k *KnowledgeManager) SearchKnowledge(knowledge_id string, text string, score float64, limit int) ([]*provider.VectorData, error) {
	if len(text) == 0 {
		return nil, errors.New("检索内容不能为空")
	}

	knowledge, err := k.GetKnowledgeInfo(knowledge_id)
	if err != nil {
		return nil, err
	}
	if len(knowledge.Embedding) == 0 {
		return nil, errors.New("向量模型未设置")
	}
	embedding := provider.CreateEmbedding(knowledge.Embedding)
	if embedding == nil {
		return nil, errors.New("向量模型不存在")
	}

	if len(knowledge.Vectordb) == 0 {
		return nil, errors.New("向量数据库未设置")
	}
	vectordb := provider.CreateVectorDB(knowledge.Vectordb)
	if vectordb == nil {
		return nil, errors.New("向量数据库不存在")
	}

	vectors, err := embedding.Embed(knowledge.EmbeddingParams, []string{text})

	if err != nil {
		return nil, errors.New("执行Embedding失败:" + err.Error())
	}
	if len(vectors) == 0 {
		return nil, errors.New("执行Embedding失败")
	}
	if score <= 0 {
		score = knowledge.Score
	}
	if limit <= 0 {
		limit = knowledge.Limit
	}

	if limit == 0 {
		limit = 5
	}

	datas, err := vectordb.Search(knowledge.VectordbParams, vectors[0], score, limit)

	return datas, err
}

// 获取文件数量
func (k *KnowledgeManager) GetAllFileCount() int {

	infos, err := k.KnowledgeInfoList()
	if err != nil {
		return 0
	}

	count := 0
	for _, info := range infos {
		count += k.GetFileCount(info.Id)

	}
	return count
}

// 获取知识集合下的文件数量
func (k *KnowledgeManager) GetFileCount(knowledge_id string) int {
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return 0
	}

	list, err := os.ReadDir(dir)
	if err != nil {

		return 0
	}

	return len(list)
}

// 知识集合下的文件列表
func (k *KnowledgeManager) GetFileInfos(knowledge_id string) ([]*meta.FileInfo, error) {
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	list, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	files := make([]*meta.FileInfo, 0)
	for _, item := range list {
		if item.IsDir() {
			info, err := k.GetFileInfo(knowledge_id, item.Name())

			if err != nil || info == nil {
				continue
			}

			files = append(files, info)
		}

	}

	return files, nil
}

// 设置文件信息
func (k *KnowledgeManager) SetFileInfo(info *meta.FileInfo) error {
	knowledge_id := info.KnowledgeId
	file_name := info.FileName
	data, err := yaml.Marshal(info)
	if err != nil {
		return err
	}
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id, file_name)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	yamlfile := path.Join(dir, "info.yaml")

	err = os.WriteFile(yamlfile, data, os.ModePerm)

	return err
}

// 获取文件信息
func (k *KnowledgeManager) GetFileInfo(knowledge_id string, file_name string) (*meta.FileInfo, error) {

	yamlfile := path.Join(k.GetKnowledgeDir(), knowledge_id, file_name, "info.yaml")
	data, err := os.ReadFile(yamlfile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}
	var info meta.FileInfo

	err = yaml.Unmarshal(data, &info)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	return &info, nil
}

// 获取文件路径
func (k *KnowledgeManager) GetFilePath(knowledge_id string, name string, ext string) string {

	filepath := path.Join(k.GetKnowledgeDir(), knowledge_id, name, "source"+ext)
	return filepath
}

// 文件上传
func (k *KnowledgeManager) FileUpload(knowledge_id string, name string, ext string, reader io.Reader) error {

	dir := path.Join(k.GetKnowledgeDir(), knowledge_id, name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	//源文件
	filepath := path.Join(dir, "source"+ext)
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	size := int64(len(b))

	word_count := len(strings.Split(string(b), ""))

	err = os.WriteFile(filepath, b, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	//文件信息
	info := &meta.FileInfo{}
	info.KnowledgeId = knowledge_id
	info.FileName = name
	info.FileExt = ext
	info.Size = size
	info.WordCount = word_count
	info.CreateTime = time.Now().Format("2006-01-02 03:04:05")
	info.ChunkType = meta.CHUNK_TYPE_NONE

	err = k.SetFileInfo(info)

	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	//更新Knowledge
	knowledge_info, _ := k.GetKnowledgeInfo(knowledge_id)
	if knowledge_info != nil {
		infos, _ := k.GetFileInfos(knowledge_id)
		count := 0
		word_count := 0
		payload_count := 0
		if infos != nil {
			count = len(infos)

			for _, info := range infos {
				word_count += info.WordCount
				payload_count += info.PayloadCount
			}
		}
		knowledge_info.FileCount = count
		knowledge_info.WordCount = word_count
		knowledge_info.PayloadCount = payload_count
		k.SetKnowledgeInfo(knowledge_info)
	}

	return nil
}

// 文件下载
func (k *KnowledgeManager) FileDownload(knowledge_id string, name string, ext string, writer io.Writer) error {

	filepath := path.Join(k.GetKnowledgeDir(), knowledge_id, name, "source"+ext)

	b, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	_, err = writer.Write(b)
	return err

}

// 删除文件
func (k *KnowledgeManager) FileRemove(knowledge_id string, filename string) error {
	lastIndex := strings.LastIndex(filename, ".")
	name := filename
	if lastIndex > 0 {
		name = filename[:lastIndex]
	}

	filepath := path.Join(k.GetKnowledgeDir(), knowledge_id, name)
	err := os.RemoveAll(filepath)

	if err != nil {
		return err
	}

	//更新Knowledge
	knowledge_info, _ := k.GetKnowledgeInfo(knowledge_id)
	if knowledge_info != nil {
		infos, _ := k.GetFileInfos(knowledge_id)
		count := 0
		word_count := 0
		payload_count := 0
		if infos != nil {
			count = len(infos)

			for _, info := range infos {
				word_count += info.WordCount
				payload_count += info.PayloadCount
			}
		}
		knowledge_info.FileCount = count
		knowledge_info.WordCount = word_count
		knowledge_info.PayloadCount = payload_count
		k.SetKnowledgeInfo(knowledge_info)
	}

	return nil
}

// 重命名文件
func (k *KnowledgeManager) FileRename(knowledge_id string, oldname string, newname string) error {

	//update dir
	olddir := path.Join(k.GetKnowledgeDir(), knowledge_id, oldname)
	newdir := path.Join(k.GetKnowledgeDir(), knowledge_id, newname)
	err := os.Rename(olddir, newdir)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	//update info
	info, err := k.GetFileInfo(knowledge_id, newname)
	if err == nil && info != nil {
		info.FileName = newname
		k.SetFileInfo(info)
	}

	return err
}

// 生成文件集合数据
func (k *KnowledgeManager) GenerateFilePayloads(knowledge_id string, file_name string) (int, error) {
	var err error

	fileInfo, err := k.GetFileInfo(knowledge_id, file_name)
	if err != nil {
		return 0, err
	}

	dir := path.Join(k.GetKnowledgeDir(), knowledge_id, file_name)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return 0, err
	}

	// 分割文件
	parts, err := k.SplitFile(fileInfo)
	if err != nil {
		return 0, err
	}

	payloads := make([]meta.Payload, 0)

	for index, word := range parts {

		//替换字符
		if len(fileInfo.WordReplace) > 0 {

			for k, v := range fileInfo.WordReplace {
				regx, err := regexp.Compile(k)
				if err == nil {
					word = string(regx.ReplaceAll([]byte(word), []byte(v)))
				} else {
					word = strings.ReplaceAll(word, k, v)
				}
			}
		}

		payloads = append(payloads, meta.Payload{Index: index, KnowledgeId: knowledge_id, FileName: file_name, Text: word, CreateTime: time.Now().Format("2006-01-02 03:04:05")})
	}

	data, err := yaml.Marshal(payloads)
	if err != nil {
		return 0, err
	}

	filepath := path.Join(dir, "payload.yaml")
	err = os.WriteFile(filepath, data, os.ModePerm)
	if err != nil {
		return 0, err
	}

	fileinfo, _ := k.GetFileInfo(knowledge_id, file_name)
	if fileinfo != nil {
		fileinfo.PayloadCount = len(payloads)
		k.SetFileInfo(fileinfo)
	}

	return fileinfo.PayloadCount, nil
}

// 分割文件
func (k *KnowledgeManager) SplitFile(fileInfo *meta.FileInfo) ([]string, error) {
	var err error
	if fileInfo.ChunkSize <= 0 {
		fileInfo.ChunkSize = 99999999999999
	}
	if fileInfo.ChunkStep <= 0 {
		fileInfo.ChunkStep = fileInfo.ChunkSize
	}

	filepath := k.GetFilePath(fileInfo.KnowledgeId, fileInfo.FileName, fileInfo.FileExt)

	// 读取文件内容
	words, err := utils.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	chunks := []string{}

	// 滑动窗口切片
	if fileInfo.ChunkType == meta.CHUNK_TYPE_WINDOW {
		word_list := strings.Split(words, "")
		for start := 0; start < len(word_list); start = start + fileInfo.ChunkStep {
			end := start + fileInfo.ChunkSize

			if end > len(word_list) {
				end = len(word_list)
			}

			word_sub := word_list[start:end]
			word := strings.Join(word_sub, "")
			word = strings.Trim(word, " ")

			chunks = append(chunks, word)
		}
	} else if fileInfo.ChunkType == meta.CHUNK_TYPE_NONE {
		chunks = append(chunks, words)
	} else if fileInfo.ChunkType == meta.CHUNK_TYPE_SPLIT {

		word_arr := strings.Split(words, fileInfo.ChunkSplit)

		chunks = append(chunks, word_arr...)
	}

	return chunks, nil
}

// 获取集合数据
func (k *KnowledgeManager) GetFilePayloads(knowledge_id string, file_name string) ([]meta.Payload, error) {
	dir := path.Join(k.GetKnowledgeDir(), knowledge_id, file_name, "payload.yaml")

	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}

	var payloads []meta.Payload
	err = yaml.Unmarshal(data, &payloads)

	return payloads, err
}

// 获取集合数据
func (k *KnowledgeManager) GetFilePayload(knowledge_id string, name string, index int) (*meta.Payload, error) {
	payloads, err := k.GetFilePayloads(knowledge_id, name)
	if err != nil {
		return nil, err
	}
	for _, payload := range payloads {
		if payload.Index == index {
			return &payload, nil
		}
	}

	return nil, nil
}
