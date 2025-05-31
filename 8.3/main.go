package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//генерируем случайную строку
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//функция небезопасного копирования
func Clone(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	return *(*string)(unsafe.Pointer(&b)) //здесь мы указываем Golang что указатель должен ссылаться на новый участок памяти, чтобы избежать автооптимизации
}

func main() {
	rand.Seed(time.Now().UnixNano())
	dataHolder := make([]string, 0)
	filler := RandStringRunes(1048576)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)

	run := true

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		run = false
		os.Exit(1)
	}()

	for {
		if !run {
			break
		}
		dataHolder = append(dataHolder, Clone(filler)) //последовательно увеличиваем обьем памяти потребляемый приложением для его выключения
		time.Sleep(1 * time.Second)
	}
}
