package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"os"

	"github.com/Dushusir/univer-server-simple/model"
	"github.com/Dushusir/univer-server-simple/utils"
	"github.com/joho/godotenv"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Document struct {
	ID     uint   `gorm:"primary_key"`
	DocsId string `json:"docs_id"`
	Type   string `json:"type"`
	Config string `json:"config"`
}

type Request struct {
	Type     string `json:"type"`
	Template string `json:"template"`
}

// wsConnection represents a WebSocket connection
type wsConnection struct {
	wsSocket            *websocket.Conn  // underlying WebSocket
	CloseChan           chan byte        // close notification
	SendChan            chan []byte      // send message
	ClientName          string           `json:"clientName"`
	SelectionActionData ISheetActionData `json:"selectionActionData"`
}

// CollaborativeDocument represents a document with collaborative data
type CollaborativeDocument struct {
	DocsId     string          `json:"docs_id"`
	DocsConfig IWorkbookConfig `json:"docsConfig"`
	Mutex      sync.Mutex
	Clients    map[string]*wsConnection `json:"-"`
	Revision   int64                    `json:"revision"`
	Data       string                   `json:"data"`
	CloseChan  chan byte                // 关闭通知
}

type Clients struct {
	ClientName          string           `json:"clientName"`
	ClientID            string           `json:"clientID"`
	SelectionActionData ISheetActionData `json:"selectionActionData"`
}
type MessageType struct {
	Type       string    `json:"type"`
	Revision   int64     `json:"revision"`
	Data       string    `json:"data"`
	ClientName string    `json:"clientName"`
	ClientID   string    `json:"clientID"`
	Clients    []Clients `json:"clients"`
}

type LocaleType string

const (
	LocaleTypeEN LocaleType = "en"
	LocaleTypeZH LocaleType = "zh"
)

type BooleanNumber int

const (
	BooleanNumberFALSE BooleanNumber = 0
	BooleanNumberTRUE  BooleanNumber = 1
)

type IWorkbookConfig struct {
	CreatedTime    string                      `json:"createdTime"`
	LastModifiedBy string                      `json:"lastModifiedBy"`
	ID             string                      `json:"id"`
	Locale         LocaleType                  `json:"locale"`
	Name           string                      `json:"name"`
	Theme          string                      `json:"theme"`
	Skin           string                      `json:"skin"`
	ModifiedTime   string                      `json:"modifiedTime"`
	TimeZone       string                      `json:"timeZone"`
	Creator        string                      `json:"creator"`
	AppVersion     string                      `json:"appVersion"`
	SocketURL      string                      `json:"socketUrl"`
	SocketEnable   BooleanNumber               `json:"socketEnable"`
	Extensions     []interface{}               `json:"extensions"`
	Styles         IKeyType[any]               `json:"styles"`
	Sheets         map[string]IWorksheetConfig `json:"sheets"`
	SheetOrder     []string                    `json:"sheetOrder"`
	PluginMeta     map[string]interface{}      `json:"pluginMeta"`
	NamedRanges    []INamedRange               `json:"namedRanges"`
}

type SheetTypes int

const (
	GRID   SheetTypes = 0
	KANBAN SheetTypes = 1
	GANTT  SheetTypes = 2
)

type IWorksheetConfig struct {
	Type               SheetTypes                             `json:"type"`
	Id                 string                                 `json:"id"`
	Name               string                                 `json:"name"`
	TabColor           string                                 `json:"tabColor"`
	Hidden             BooleanNumber                          `json:"hidden"`
	FreezeRow          int                                    `json:"freezeRow"`
	FreezeColumn       int                                    `json:"freezeColumn"`
	RowCount           int                                    `json:"rowCount"`
	ColumnCount        int                                    `json:"columnCount"`
	ZoomRatio          float64                                `json:"zoomRatio"`
	ScrollTop          float64                                `json:"scrollTop"`
	ScrollLeft         float64                                `json:"scrollLeft"`
	DefaultColumnWidth int                                    `json:"defaultColumnWidth"`
	DefaultRowHeight   int                                    `json:"defaultRowHeight"`
	MergeData          []IRangeData                           `json:"mergeData"`
	HideRow            []int                                  `json:"hideRow"`
	HideColumn         []int                                  `json:"hideColumn"`
	Status             BooleanNumber                          `json:"status"`
	CellData           ObjectMatrixPrimitiveType[interface{}] `json:"cellData"`
	RowData            interface{}                            `json:"rowData"`
	ColumnData         interface{}                            `json:"columnData"`
	ShowGridlines      BooleanNumber                          `json:"showGridlines"`
	RowTitle           struct {
		Width  int           `json:"width"`
		Hidden BooleanNumber `json:"hidden,omitempty"`
	} `json:"rowTitle"`
	ColumnTitle struct {
		Height int           `json:"height"`
		Hidden BooleanNumber `json:"hidden,omitempty"`
	} `json:"columnTitle"`
	Selections  []interface{}          `json:"selections"`
	RightToLeft BooleanNumber          `json:"rightToLeft"`
	PluginMeta  map[string]interface{} `json:"pluginMeta"`
}

type IKeyType[T any] map[string]T

type INamedRange struct {
	NamedRangeID string     `json:"namedRangeId"`
	Name         string     `json:"name"`
	Range        IGridRange `json:"range"`
}

type IGridRange struct {
	SheetID   string     `json:"sheetId"`
	RangeData IRangeData `json:"rangeData"`
}

type IRangeData struct {
	StartRow    int `json:"startRow"`
	StartColumn int `json:"startColumn"`
	EndRow      int `json:"endRow"`
	EndColumn   int `json:"endColumn"`
}

type MessageConfigType struct {
	UniverId   string           `json:"univerId"`
	ActionData ISheetActionData `json:"actionData"`
}

type ISheetActionData struct {
	IActionData
	ISetRangeDataActionData
	ISetSelectionValueActionData
	SheetId string `json:"sheetId"`
}

type IActionData struct {
	ActionName string `json:"actionName"`
	Operation  int    `json:"operation"`
}

// action
type ISetRangeDataActionData struct {
	CellValue ObjectMatrixPrimitiveType[interface{}] `json:"cellValue"`
}

type ObjectMatrixPrimitiveType[T any] map[string]map[string]T

type ISetSelectionValueActionData struct {
	Selections []ISelectionModelValue `json:"selections"`
}

type ISelectionModelValue struct {
	Selection ISelection `json:"selection"`
	Cell      ICellInfo  `json:"cell"`
}

type ISelection struct {
	StartX      int `json:"startX"`
	StartY      int `json:"startY"`
	EndX        int `json:"endX"`
	EndY        int `json:"endY"`
	StartRow    int `json:"startRow"`
	StartColumn int `json:"startColumn"`
	EndRow      int `json:"endRow"`
	EndColumn   int `json:"endColumn"`
}

type ICellInfo struct {
	StartX           int        `json:"startX"`
	StartY           int        `json:"startY"`
	EndX             int        `json:"endX"`
	EndY             int        `json:"endY"`
	Row              int        `json:"row"`
	Column           int        `json:"column"`
	IsMerged         bool       `json:"isMerged"`
	IsMergedMainCell bool       `json:"isMergedMainCell"`
	MergeInfo        ISelection `json:"mergeInfo"`
}

// Upgrader is used to upgrade HTTP connection to WebSocket connection
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024 * 1024,
	WriteBufferSize: 1024 * 1024 * 1024,
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// DocMap stores CollaborativeDocument objects by key
var DocMap = make(map[string]*CollaborativeDocument)

func main() {

	// // 生成10个不重复的名字
	// for i := 0; i < 100; i++ {
	// 	name := utils.GetUniqueName()
	// 	fmt.Println("GET name====",name.First, name.Last)
	// }

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	DbHost := os.Getenv("DB_HOST")
	DbUser := os.Getenv("DB_USER")
	DbPassword := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")
	DbPort := os.Getenv("DB_PORT")

	DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort)

	DBURL_NAME := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(DBURL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 创建数据库
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS " + DbName).Error; err != nil {
		panic("failed to create database")
	}
	// 连接到数据库
	db, err = gorm.Open(mysql.Open(DBURL_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	db.AutoMigrate(&Document{})

	// 挂载到model中，在其他模块中就可以使用了
	model.DB = db

	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.Use(cors.Default())

	// test with POST localhost:8080/new, application/json {"type":"sheet","template":"DEMO1"}
	router.POST("/new", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var config string

		switch req.Type {
		case "sheet":
			// TODO maybe json string
			config = "default"
		case "doc":
			config = "default"
		case "slide":
			config = "default"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
			return
		}

		docs_id := utils.GenerateID(10)

		document := Document{
			Type:   req.Type,
			DocsId: docs_id,
			Config: config,
		}
		db.Create(&document)

		c.JSON(http.StatusOK, gin.H{"type": document.Type, "id": document.DocsId, "config": document.Config})
	})

	// 定义打开文档的 POST 接口
	router.POST("/open", func(c *gin.Context) {
		// 获取请求参数
		docs_id := c.PostForm("id")
		fmt.Println("id==", docs_id)
		// 查询数据库，获取文档信息
		var document Document

		if err := db.Where("docs_id = ?", docs_id).First(&document).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Document not found"})
			return
		}

		// 从当前协同的内存中取最新document数据
		doc, _ := DocMap[docs_id]

		if doc != nil {
			docsConfig, _ := json.Marshal(doc.DocsConfig)
			document.Config = string(docsConfig)

			fmt.Println("document open from collaboration", document)
		}

		// 返回文档信息
		c.JSON(http.StatusOK, gin.H{"type": document.Type, "id": document.DocsId, "config": document.Config})
	})

	// 实现更新文档配置的接口
	router.POST("/update", func(c *gin.Context) {
		// 获取请求参数
		docs_id := c.PostForm("id")
		config := c.PostForm("config")

		// 查询文档记录
		var document Document
		if err := db.Where("docs_id = ?", docs_id).First(&document).Error; err != nil {
			c.AbortWithError(400, err)
			return
		}

		// 更新文档配置
		document.Config = config
		if err := db.Save(&document).Error; err != nil {
			c.AbortWithError(500, err)
			return
		}

		// 返回更新后的文档信息
		c.JSON(200, document)
	})

	// Handle websocket connections
	router.GET("/ws/:id", handleWebSocket)

	fmt.Println("Start server start at: localhost:8080")
	router.Run(":8500")
}

func handleWebSocket(c *gin.Context) {
	// Get the document key from the request URL
	docs_id := c.Param("id")

	var db = model.DB

	// Check if the document exists in the database
	var document Document
	result := db.Where("docs_id = ?", docs_id).First(&document)
	if result.Error != nil {
		c.AbortWithError(http.StatusNotFound, result.Error)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Create a CollaborativeDocument object if it doesn't exist
	doc, ok := DocMap[docs_id]
	if !ok {
		var docsConfig IWorkbookConfig
		err = json.Unmarshal([]byte(document.Config), &docsConfig)

		doc = &CollaborativeDocument{
			DocsId:     docs_id,
			DocsConfig: docsConfig,
			Clients:    make(map[string]*wsConnection),
			Revision:   0,
			Data:       "Initialize collaborative data",
		}
		DocMap[docs_id] = doc
	}

	// Generate a unique client ID
	clientID := "user-" + utils.GenerateID(10)
	clientName := utils.GetUniqueName()

	// Create a wsConnection object for the client
	wsConn := &wsConnection{
		wsSocket:   conn,
		CloseChan:  make(chan byte),
		SendChan:   make(chan []byte, 256),
		ClientName: clientName.First,
	}

	// Add the client to the list of clients for the document
	doc.Mutex.Lock()
	doc.Clients[clientID] = wsConn
	doc.Mutex.Unlock()

	// Send the document data to the client
	sendDocumentData(clientID, doc, wsConn)

	// Start the client read loop
	go readLoop(clientID, doc, wsConn)

	// Start the client write loop
	go writeLoop(clientID, doc, wsConn)
}

func sendDocumentData(clientID string, doc *CollaborativeDocument, wsConn *wsConnection) {
	// Send the document data to the client
	message := MessageType{
		Type:     "document",
		Revision: doc.Revision,
		Data:     doc.Data,
		ClientID: clientID,
	}

	err := wsConn.wsSocket.WriteJSON(message)
	if err != nil {
		log.Println(err)
		return
	}

	// Send the list of clients to the client
	clientList := []Clients{}
	doc.Mutex.Lock()
	for id, c := range doc.Clients {
		clientList = append(clientList, Clients{ClientName: c.ClientName, ClientID: id, SelectionActionData: c.SelectionActionData})
	}
	doc.Mutex.Unlock()
	userMessage := MessageType{
		Type:    "clients",
		Clients: clientList,
	}
	err = wsConn.wsSocket.WriteJSON(userMessage)
	if err != nil {
		log.Println(err)
		return
	}
}

func readLoop(clientID string, doc *CollaborativeDocument, wsConn *wsConnection) {
	// Close the connection when this function exits
	conn := wsConn.wsSocket
	defer conn.Close()

	// Loop to read messages from the client
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			// Remove the client from the list of clients for the document
			doc.Mutex.Lock()
			delete(doc.Clients, clientID)
			offline(clientID, wsConn.ClientName, doc)
			doc.Mutex.Unlock()

			log.Println("Client", clientID, "disconnected from document", doc.DocsId)
			clients := len(doc.Clients)
			log.Println("Read Message Current clients number: ", clients)

			if clients == 0 {
				saveDocument(doc)
			}

			return
		}

		// Parse the message
		var message MessageType
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		// Handle the message based on its type
		switch message.Type {
		case "data":
			// Update the collaborative data for the document
			doc.Mutex.Lock()
			doc.Data = message.Data

			// update to current config
			updateDocument(doc, wsConn)

			doc.Revision++
			doc.Mutex.Unlock()

			// Update the database with the new collaborative data
			// updateCollaborativeData(doc.DocsId, doc.Data)

			// Broadcast the new collaborative data to all clients for the document
			doc.Mutex.Lock()
			for _, currentWsConn := range doc.Clients {
				err := currentWsConn.wsSocket.WriteJSON(MessageType{
					Type:       "data",
					Revision:   doc.Revision,
					Data:       doc.Data,
					ClientID:   clientID,
					ClientName: wsConn.ClientName,
				})
				if err != nil {
					log.Println("Error sending message to client:", err)
				}
			}
			doc.Mutex.Unlock()
		default:
			log.Println("Invalid message type:", message.Type)
		}
	}
}

func updateDocument(doc *CollaborativeDocument, wsConn *wsConnection) {
	// 更新文档配置
	// document := doc.Document
	var data MessageConfigType
	err := json.Unmarshal([]byte(doc.Data), &data)

	if err != nil {
		log.Println("Error parsing message:", err)
	}

	var config = doc.DocsConfig

	if err != nil {
		log.Println("Error parsing config:", err)
	}

	sheet := config.Sheets[data.ActionData.SheetId]

	actionData := data.ActionData

	// 只处理需要落库的数据，比如更新单元格需要更新，但是选区高亮只协同不更新数据库
	switch actionData.ActionName {
	case "SetRangeDataAction":
		updateCellData(&sheet, &actionData)

	case "SetSelectionValueAction":
		updateSelction(wsConn, &actionData)
	}
}

func saveDocument(doc *CollaborativeDocument) {

	db := model.DB
	var document Document
	// 更新文档配置
	docsConfig, _ := json.Marshal(doc.DocsConfig)

	if err := db.Model(&document).Where("docs_id = ?", doc.DocsId).Update("config", string(docsConfig)).Error; err != nil {
		panic("failed to update document")
	}

	fmt.Println("update document success")
}

func updateCellData(sheet *IWorksheetConfig, actionData *ISheetActionData) ObjectMatrixPrimitiveType[interface{}] {
	for row, rowData := range actionData.CellValue {
		if sheet.CellData[row] == nil {
			sheet.CellData[row] = make(map[string]interface{})
		}
		for column, value := range rowData {
			sheet.CellData[row][column] = value
		}
	}
	return sheet.CellData
}
func updateSelction(wsConn *wsConnection, actionData *ISheetActionData) {
	wsConn.SelectionActionData = *actionData
}

func writeLoop(clientID string, doc *CollaborativeDocument, wsConn *wsConnection) {
	// Close the connection when this function exits
	conn := wsConn.wsSocket
	defer conn.Close()

	// Loop to write messages to the client
	for {
		// Wait for a signal to send the current document state to the client
		select {
		case <-wsConn.CloseChan:
			// The client disconnected, so remove it from the list of clients for the document
			doc.Mutex.Lock()
			delete(doc.Clients, clientID)
			offline(clientID, wsConn.ClientName, doc)
			doc.Mutex.Unlock()

			log.Println("Client", clientID, "disconnected from document", doc.DocsId)

			clients := len(doc.Clients)
			log.Println("CloseChan Current clients number: ", clients)

			if clients == 0 {
				saveDocument(doc)
			}

			return
		case <-wsConn.SendChan:
			// Send the current document state to the client
			err := conn.WriteJSON(MessageType{
				Type:     "data",
				Revision: doc.Revision,
				Data:     doc.Data,
			})
			if err != nil {
				log.Println("Error sending message to client:", err)
			}
		}
	}
}

// 下线通知
func offline(clientID string, clientName string, doc *CollaborativeDocument) {

	for _, currentWsConn := range doc.Clients {
		err := currentWsConn.wsSocket.WriteJSON(MessageType{
			Type:       "offline",
			ClientID:   clientID,
			ClientName: clientName,
		})
		if err != nil {
			log.Println("Error sending message to client:", err)
		}
	}
}
