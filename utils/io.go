package utils

import (
	"errors"
	"os"
	"runtime"
	"sync"
)

// IsExist returns whether the given file or directory exists
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		//log.Println(filepath.Join(path), "найден!")
		return true
	}
	if os.IsNotExist(err) {
		//log.Println(filepath.Join(path), "не найден")
		return false
	}
	//log.Println(filepath.Join(path), "[?] Что я здесь делаю?")
	return false
}

// EnsureMakeDir гарантированное создание вложенных каталогов на диске
func EnsureMakeDir(fileName string) error {

	if _, serr := os.Stat(fileName); serr != nil {
		err := os.MkdirAll(fileName, os.ModePerm) //err := os.MkdirAll(dirName, os.ModePerm)
		if os.IsExist(err) {
			//log.Println("[i] Папка уже существует")
			return nil
		}
		if err == nil {
			//log.Println("[i] Папка создана")
			return nil
		} else {
			//log.Errorln("error folder creating", err)
			return err
		}
	}
	return nil
}

// UserHomeDir возвращает домашнюю папку юзера
func UserHomeDir() string {

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

type SafeFile struct {
	*os.File
	lock sync.Mutex
}

func (sf *SafeFile) WriteAt(b []byte, off int64) (n int, err error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	return sf.File.WriteAt(b, off)
}

func (sf *SafeFile) Sync() error {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	return sf.File.Sync()
}
func OpenSafeFile(name string) (file *SafeFile, err error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0666)
	return &SafeFile{File: f}, err
}

func CreateSafeFile(name string) (file *SafeFile, err error) {
	if IsExist(name) {
		return nil, errors.New("file exist")
	}
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	return &SafeFile{File: f}, err
}

func getFileSize(filepath string) (int64, error) {
	var fileSize int64
	fi, err := os.Stat(filepath)
	if err != nil {
		return fileSize, err
	}
	if fi.IsDir() {
		return fileSize, nil
	}
	fileSize = fi.Size()
	return fileSize, nil
}
