package main 

import (
	"html/template"
	"log"
	"net/http"
	"database/sql"
	"strconv"
    
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type indexPage struct {
	Title         string
	FeaturedPosts []featuredPostData
	MostRecent    []mostRecentData
}

type postPage struct {
	Title       string `db:"title"`
	Subtitle    string `db:"subtitle"`
	ImgModifier string `db:"image_url"`
	Content     string `db:"content"`
}

type featuredPostData struct {
	PostID      string `db:"post_id"`
	Title       string `db:"title"`
	Subtitle    string `db:"subtitle"`
	ImgModifier string `db:"image_url"`
	Author      string `db:"author"`
	AuthorImg   string `db:"author_url"`
	PublishDate string `db:"publish_date"`
 }

 type mostRecentData struct {
	PostID      string `db:"post_id"`
	Title       string `db:"title"`
	Subtitle    string `db:"subtitle"`
	ImgModifier string `db:"image_url"`
	Author      string `db:"author"`
	AuthorImg   string `db:"author_url"`
	PublishDate string `db:"publish_date"`
 }

func index(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mostRecent, err := mostRecent(db)
		featuredPosts, err := featuredPosts(db)	
		if err != nil {
			http.Error(w, "Internal Server Error", 500) // В случае ошибки парсинга - возвращаем 500
			log.Println(err)
			return // Не забываем завершить выполнение ф-ии
		}

		ts, err := template.ParseFiles("pages/index.html") // Главная страница блога
		if err != nil {
			http.Error(w, "Internal Server Error", 500) // В случае ошибки парсинга - возвращаем 500
			log.Println(err)
			return // Не забываем завершить выполнение ф-ии
		}

		data := indexPage{
			Title:         "Escape",
			FeaturedPosts: featuredPosts,
			MostRecent: mostRecent,
		}

		err = ts.Execute(w, data) // Заставляем шаблонизатор вывести шаблон в тело ответа
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		log.Println("Request completed successfully")
	}
}

func featuredPosts(db *sqlx.DB) ([]featuredPostData, error) {
	const query = `
		SELECT
		    post_id,
			title,
			subtitle,
			author,
			author_url,
			publish_date,
			image_url
		FROM
			post
		WHERE featured = 1
		LIMIT 2
	` // Составляем SQL-запрос для получения записей для секции featured-posts

	var featuredPosts []featuredPostData // Заранее объявляем массив с результирующей информацией

	err := db.Select(&featuredPosts, query) // Делаем запрос в базу данных
	if err != nil {                 // Проверяем, что запрос в базу данных не завершился с ошибкой
		return nil, err
	}

	return featuredPosts, nil
}
func mostRecent(db *sqlx.DB) ([]mostRecentData, error) {
	const query = `
		SELECT
		    post_id,
			title,
			subtitle,
			author,
			author_url,
			publish_date,
			image_url
		FROM
			post
		WHERE featured = 0
		ORDER BY post_id DESC
		LIMIT 6
	` // Составляем SQL-запрос для получения записей для секции featured-posts

	var mostRecent []mostRecentData// Заранее объявляем массив с результирующей информацией

	err := db.Select(&mostRecent, query) // Делаем запрос в базу данных
	if err != nil {                 // Проверяем, что запрос в базу данных не завершился с ошибкой
		return nil, err
	}

	return mostRecent, nil
}

func post(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := mux.Vars(r)["postID"] // Получаем orderID в виде строки из параметров урла

		postID, err := strconv.Atoi(postIDStr) // Конвертируем строку orderID в число
		if err != nil {
			http.Error(w, "Invalid post id", 403)
			log.Println(err)
			return
		}

		post, err := postByID(db, postID)
		if err != nil {
			if err == sql.ErrNoRows {
				// sql.ErrNoRows возвращается, когда в запросе к базе не было ничего найдено
				// В таком случае мы возвращем 404 (not found) и пишем в тело, что ордер не найден
				http.Error(w, "post not found", 404)
				log.Println(err)
				return
			}

			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		err = ts.Execute(w, post)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		log.Println("Request completed successfully")
	}
}

func postByID(db *sqlx.DB, postID int) (postPage, error) {
	const query = `
		SELECT
		    image_url,
			title,
			subtitle,
			content
		FROM
			post
		WHERE
			post_id = ?
	`
	// В SQL-запросе добавились параметры, как в шаблоне. ? означает параметр, который мы передаем в запрос ниже

	var post postPage

	// Обязательно нужно передать в параметрах orderID
	err := db.Get(&post, query, postID)
	if err != nil {
		return postPage{}, err
	}

	return post, nil
}
