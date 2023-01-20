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
