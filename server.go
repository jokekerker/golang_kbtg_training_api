package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

// func helloHandler(w http.ResponseWriter, req *http.Request) {
// 	resp := []byte(`{"name":"papatpon"}`)
// 	w.Write(resp)
// }

type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func helloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello",
	})
}

var todos = map[int]*Todo{}

func getAllTodosHandler(c *gin.Context) {

	db, err := sql.Open("postgres", "postgres://vpcujjbj:GqfnT0fLF63MZyB4YvxklDt-xhZe6aUF@suleiman.db.elephantsql.com:5432/vpcujjbj")
	if err != nil {
		log.Fatal("connect database error", err)
	}
	defer db.Close()

	queryDb := `
	select id, title, status from todos
	`

	stmt, err := db.Prepare(queryDb)

	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("can't query all todos", err)
		return
	}

	// items := []*Todo{}
	for rows.Next() {
		// var id int
		// var title, status string
		t := Todo{}
		err := rows.Scan(&t.ID, &t.Title, &t.Status)

		if err != nil {
			log.Fatal("can't scan row into variable", err)
			return
		}
		// fmt.Println("one row", id, title, status)

		todos[t.ID] = &t
	}

	// for _, item := range todos {
	// 	items = append(items, item)
	// }
	// data := []byte(`{
	//         "ID": 1,
	//         "title": "pay credit card",
	//         "status": "active"
	//     }`)

	// t := Todo{}
	// err := json.Unmarshal(data, &t)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, err)
	// 	return
	// }
	c.JSON(http.StatusOK, todos)
}

func getTodoByIdHandler(c *gin.Context) {

	db, err := sql.Open("postgres", "postgres://vpcujjbj:GqfnT0fLF63MZyB4YvxklDt-xhZe6aUF@suleiman.db.elephantsql.com:5432/vpcujjbj")
	if err != nil {
		log.Fatal("connect database error", err)
	}
	defer db.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	queryDb := `
	select id, title, status from todos where id=$1
	`

	stmt, err := db.Prepare(queryDb)

	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
		return
	}
	row := stmt.QueryRow(id)
	// t, ok := todos[id]
	t := Todo{}
	err = row.Scan(&t.ID, &t.Title, &t.Status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, t)
}

func createTodosHandler(c *gin.Context) {
	db, err := sql.Open("postgres", "postgres://vpcujjbj:GqfnT0fLF63MZyB4YvxklDt-xhZe6aUF@suleiman.db.elephantsql.com:5432/vpcujjbj")
	if err != nil {
		log.Fatal("connect database error", err)
	}
	defer db.Close()

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	insertDb := `
	insert into todos (title, status) values ($1, $2) returning id;
	`

	t := Todo{}
	err2 := json.Unmarshal(jsonData, &t)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, err2)
		return
	}

	row := db.QueryRow(insertDb, &t.Title, &t.Status)
	err = row.Scan(&t.ID)

	c.JSON(http.StatusCreated, &t)
}

func filteringByStatus(c *gin.Context) {
	status := c.DefaultQuery("status", "Guest")

	items := []*Todo{}
	for _, item := range todos {
		if item.Status == status {
			items = append(items, item)
		}
	}

	c.JSON(http.StatusOK, items)
}

func updateTodosHandler(c *gin.Context) {
	db, err := sql.Open("postgres", "postgres://vpcujjbj:GqfnT0fLF63MZyB4YvxklDt-xhZe6aUF@suleiman.db.elephantsql.com:5432/vpcujjbj")
	if err != nil {
		log.Fatal("connect database error", err)
	}
	defer db.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	t := Todo{}
	err2 := json.Unmarshal(jsonData, &t)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, err2)
		return
	}

	updateDb := `
	update todos set status=$2, title=$3 where id=$1;
	`
	stmt, err := db.Prepare(updateDb)

	if err != nil {
		log.Fatal("can't prepare statement update", err)
		return
	}

	if _, err := stmt.Exec(id, &t.Status, &t.Title); err != nil {
		log.Fatal("error execute update ", err)
	}

	c.JSON(http.StatusOK, "success")
}

func main() {
	/** gin + database ==> Todo api
	    POST /todos
	      body: {​​​​​ "status": "active", "title": "homework #2" }​​​​​
	      response: {​​​​​ id: 1, "status": "active", "title": "homework #2" }​​​​​
	    GET /todos
	      response: [{​​​​​ id: 1, "status": "active", "title": "homework #2" }​​​​​]
	    GET /todos/1
	      response: {​​​​​ id: 1, "status": "active", "title": "homework #2" }​​​​​
	   PUT /todos/1
	      body: {​​​​​ "status": "completed", "title": "homework #2" }​​​​​
	      response: {​​​​​ "status": "success" }​​​​​
	   DELETE /todos/1
	      body: {​​​​​ "status": "completed", "title": "homework #2" }​​​​​
	      response: {​​​​​ "status": "deleted" }​​​​​
	   */
	// fmt.Println("start...")
	// http.HandleFunc("/", helloHandler)
	// log.Fatal(http.ListenAndServe(":123456", nil))
	// fmt.Println("end...")

	r := gin.Default()
	r.GET("/hello", helloHandler)
	r.GET("/todos", getAllTodosHandler)
	r.GET("/todos/:id", getTodoByIdHandler)
	r.POST("/todos", createTodosHandler)
	r.PUT("/todos/:id", updateTodosHandler)
	r.Run(":1234")
}
