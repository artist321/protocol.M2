package utils

import "os"

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
