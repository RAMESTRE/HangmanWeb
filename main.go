package main

import (
	"encoding/json"
	"fmt"
	"hangmanclassic"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

func main() {
	server := http.NewServeMux()
	// url http://localhost:8000/
	server.Handle("/spriteStickman/", http.StripPrefix("/spriteStickman/", http.FileServer(http.Dir("spriteStickman"))))
	server.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	server.HandleFunc("/", MainPage)
	server.HandleFunc("/rules", Rules)
	server.HandleFunc("/difficulty", DifficultyMenu)
	server.HandleFunc("/saveDifficulty", WhatDifficulty)

	server.HandleFunc("/nickname", ChooseNickname)
	server.HandleFunc("/saveNicknames", saveNicknames)
	server.HandleFunc("/whoguess", whoGuess)
	server.HandleFunc("/saveWhoWriteWord", saveWhoWriteWord)
	server.HandleFunc("/transformWordChosen",DuoMode)
	server.HandleFunc("/duoword", Duoword)

	server.HandleFunc("/play", Playpage)
	server.HandleFunc("/hangman", Hangman)
	server.HandleFunc("/win", WinPage)
	server.HandleFunc("/lose", LosePage)
	server.HandleFunc("/restart", RestartGame)

	// listen to the port 8000
	fmt.Println("server listening on http://localhost:8080")
	http.ListenAndServe(":8080", server)
}


type BasicParameters struct {
	Attempt         int
	Word        []string
	WrongLetter bool
	WrongWord   bool
	CopyWord    []string
	GuessWord   []string
	AlreadyUsed []string
	NicknameP1  string
	NicknameP2  string
	Solo bool
	WordFile string
	NicknameChoose string
}

func GameProperties(vie int, WrongLetter, WrongWord, solo bool, word, copyWord, guessWord, alreadyUsed []string, nicknameP1, nicknameP2, wordFile, nicknameChoose string) {
	save := BasicParameters{vie, word, WrongLetter, WrongWord, copyWord, guessWord, alreadyUsed, nicknameP1, nicknameP2, solo, wordFile, nicknameChoose}
	encoded, _ := json.Marshal(save)
	err2 := ioutil.WriteFile("save.txt", encoded, 0777)
	if err2 != nil {
		fmt.Println("Error while writing in save.txt")
		os.Exit(1)
	}
	fmt.Println("Basic Property saved")
}

func ReadGameProperties() BasicParameters {
	var m BasicParameters
	content, _ := ioutil.ReadFile("save.txt")
	err2 := json.Unmarshal(content, &m)
	if err2 != nil {
		fmt.Println("Impossible to decode the json data in the file")
		os.Exit(7)
	}
	return m
}


func MainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/mainpage.html")
}

func Rules(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/rules.gohtml"))
	_ = tmpl.Execute(w, nil)
}

func DifficultyMenu(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/difficulty.gohtml"))
	_ = tmpl.Execute(w, nil)
}

func WhatDifficulty(w http.ResponseWriter, r *http.Request) {
	word, guessWord := hangmanclassic.Computer(r.FormValue("difficulty"))
	copyWord := hangmanclassic.LetterAlreadyHere(word, guessWord)
	GameProperties(10, true, true, true, word, copyWord, guessWord, []string{}, "", "", r.FormValue("difficulty"), "")
	http.Redirect(w, r, "/play", http.StatusFound)
}


func ChooseNickname(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/choosepseudo.gohtml"))
	_ = tmpl.Execute(w, nil)
}

func saveNicknames(w http.ResponseWriter, r *http.Request){
	decoded := ReadGameProperties()
	GameProperties(10, true, true, false, decoded.Word, decoded.CopyWord, decoded.GuessWord, []string{}, r.FormValue("nicknamePlayer1"), r.FormValue("nicknamePlayer2"), decoded.WordFile, decoded.NicknameChoose)
	http.Redirect(w,r,"/whoguess", http.StatusFound)
}

func whoGuess(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	tmpl := template.Must(template.ParseFiles("templates/whoguess.gohtml"))
	_ = tmpl.Execute(w, struct {
		NicknameP1 string
		NicknameP2 string
	}{NicknameP1: decoded.NicknameP1, NicknameP2: decoded.NicknameP2})
}

func saveWhoWriteWord(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	GameProperties(10, true, true, false, decoded.Word, decoded.CopyWord, decoded.GuessWord, []string{}, decoded.NicknameP1, decoded.NicknameP2, decoded.WordFile, r.FormValue("player"))
	http.Redirect(w,r, "/duoword", http.StatusFound)
}

func DuoMode(w http.ResponseWriter, r *http.Request) {
	redirect := "/play"
	decoded := ReadGameProperties()
	var word []string
	var copyWord []string
	var guessWord []string
	var listSearch []string
	wrongLetter := true
	if hangmanclassic.WordChosenRight(r.FormValue("wordtoguess")) {
		for _, element := range r.FormValue("wordtoguess") {
			word = append(word, string(element))
		}
		for i := 0; i < len(word); i++ {
			listSearch = append(listSearch, "_")
		}
		guessWord = hangmanclassic.ShowRandomLetterWord(word, listSearch)
		copyWord = hangmanclassic.LetterAlreadyHere(word, guessWord)
	} else {
		wrongLetter = false
		redirect = "/duoword"
	}
	GameProperties(decoded.Attempt, true, wrongLetter, false, word, copyWord, guessWord, []string{}, decoded.NicknameP1, decoded.NicknameP2, decoded.WordFile, decoded.NicknameChoose)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func Duoword(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	tmpl := template.Must(template.ParseFiles("templates/duowordchoice.gohtml"))
	_ = tmpl.Execute(w, struct {
		Nickname string
	}{Nickname: decoded.NicknameChoose})
}


func Playpage(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	tmpl := template.Must(template.ParseFiles("templates/index.gohtml"))
	_ = tmpl.Execute(w, struct {
		PageTitle   string
		GuessWord   []string
		UsedLetters []string
		WrongLetter bool
		WrongWord   bool
		Word        []string
		Attempt         int
		Solo bool
	}{PageTitle: "Hangman Web", GuessWord: decoded.GuessWord, Word: decoded.Word, Attempt: decoded.Attempt, UsedLetters: decoded.AlreadyUsed, WrongLetter: decoded.WrongLetter, WrongWord: decoded.WrongWord, Solo: decoded.Solo})
}

func Hangman(w http.ResponseWriter, r *http.Request) {
	redirect := "/play"
	decoded := ReadGameProperties()
	usedLetters := decoded.AlreadyUsed
	decoded.WrongLetter, decoded.WrongWord = true, true
	letter, word := r.FormValue("Letter"), r.FormValue("Word")
	var alreadyHere bool
	var successfull string

	if letter == "STOP" || word == "STOP" {
		decoded.WrongWord = false
	} else if letter != "" {
		decoded.WrongLetter = hangmanclassic.CorrectLetter(letter)
		if decoded.WrongLetter {
			alreadyHere = hangmanclassic.OutLetters(usedLetters, letter)
			if alreadyHere {
				successfull = hangmanclassic.Success(decoded.CopyWord, decoded.GuessWord, letter)
				usedLetters = append(usedLetters, letter)
				sort.Strings(usedLetters)
				decoded.Attempt = hangmanclassic.NumberOfLife(successfull, decoded.Attempt)
			}
		}
	} else if word != "" {
		if hangmanclassic.OnlyLowerLetterWord(r.FormValue("Word")) == false {
			decoded.WrongWord = false
		} else {
			successfull = hangmanclassic.CorrectWord(r.FormValue("Word"), decoded.Word)
			if successfull == "true" {
				decoded.GuessWord = decoded.Word
			} else {
				decoded.Attempt = hangmanclassic.NumberOfLife(successfull, decoded.Attempt)
			}
		}
	}
	if decoded.Attempt == 0 {
		redirect = "/lose"
	} else if hangmanclassic.Victory(decoded.GuessWord, decoded.Word) {
		redirect = "/win"
	}
	GameProperties(decoded.Attempt, decoded.WrongLetter, decoded.WrongWord, decoded.Solo, decoded.Word, decoded.CopyWord, decoded.GuessWord, usedLetters, decoded.NicknameP1, decoded.NicknameP2, decoded.WordFile, decoded.NicknameChoose)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func WinPage(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	tmpl := template.Must(template.ParseFiles("templates/win.gohtml"))
	_ = tmpl.Execute(w, struct {
		PageTitle string
		Word      []string
		Solo bool
		NicknameP1 string
		NicknameP2 string
		NicknameChoose string
	}{PageTitle: "Hangman Web", Word: decoded.Word, Solo: decoded.Solo, NicknameP1: decoded.NicknameP1, NicknameP2: decoded.NicknameP2, NicknameChoose: decoded.NicknameChoose})
}

func LosePage(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	tmpl := template.Must(template.ParseFiles("templates/lose.gohtml"))
	_ = tmpl.Execute(w, struct {
		PageTitle string
		Word      []string
		Solo bool
		NicknameP1 string
		NicknameP2 string
		NicknameChoose string
	}{PageTitle: "Hangman Web", Word: decoded.Word, Solo: decoded.Solo, NicknameP1: decoded.NicknameP1, NicknameP2: decoded.NicknameP2, NicknameChoose: decoded.NicknameChoose})
}

func RestartGame(w http.ResponseWriter, r *http.Request) {
	decoded := ReadGameProperties()
	word, guessWord := hangmanclassic.Computer(decoded.WordFile)
	copyWord := hangmanclassic.LetterAlreadyHere(word, guessWord)
	GameProperties(10, true, true, true, word, copyWord, guessWord, []string{}, "", "", decoded.WordFile, "")
	http.Redirect(w, r, "/play", http.StatusFound)
}

