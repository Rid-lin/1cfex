// main project main.go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/jlaffaye/ftp"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var configExample = `# Порядок не имеет значения
login_FTP = kust
pass_FTP = qazwsx
Path = /srv/1cv8/uat/
local_path = C:/ftpswap/LocalObmenUAT/
server = 10.57.254.103:21
file_in = Message_КРК_ТСТ.zip
file_out = Message_ТСТ_КРК.zip`

var err error
var c *ftp.ServerConn
var configFileName = "1cfex.ini"
var codeExit int

//ConfigAttr - configuration attribute
type ConfigAttr struct {
	ServerPort string
	Login      string
	Password   string
	Path       string
	LocalPath  string
	FileIn     string
	FileOut    string
}

//Servers - list of servers for check
var cfg ConfigAttr

func intro() {
	fmt.Printf(`
                           1C File eXchange for 
    "Management of Technological Transport & Special Mechanism Burservice"
                            Copyright (C) 2018
1cfex 1.0.1                     Vlad Vegner                     July 16th 2018
===============================================================================
`)
}

//LoadConfig - Load config
func (s *ConfigAttr) LoadConfig(nameFile string) {
	fmt.Printf("%v Загружаю список серверов из %v", time.Now().Format("15:04:05"), nameFile)
	cfg, err := ini.Load(nameFile)
	if err != nil {
		fmt.Printf("\nОшибка чтения конфигурационного файла: %v", err)
		os.Exit(1)
	}
	fmt.Printf("\t...\tОК\n")
	s.ServerPort = cfg.Section("").Key("server").String()
	s.Login = cfg.Section("").Key("login_FTP").String()
	s.Password = cfg.Section("").Key("pass_FTP").String()
	s.Path = cfg.Section("").Key("path").String()
	s.LocalPath = cfg.Section("").Key("local_path").String()
	s.FileIn = cfg.Section("").Key("file_in").String()
	s.FileOut = cfg.Section("").Key("file_out").String()
}

// ConnectToFTP Connect, Login, change DIR
func ConnectToFTP(ServerPort, Login, Password, Path string) (c *ftp.ServerConn) {
	// Подключаюсь к серверу
	fmt.Printf("%v Устанавливаю соедиение с сервером %s, авторизуюсь, меняю папку ", time.Now().Format("15:04:05"), ServerPort)
	c, err = ftp.Dial(ServerPort)
	if err != nil {
		fmt.Printf("не удалось\n %v \n", err)
		os.Exit(1) // Если не удалось соединитьяс с сервером, то выходим с кодом
	} else {
		fmt.Printf("\t.")
	}
	// Ввожу логин и пароль
	if err := c.Login(Login, Password); err != nil {
		fmt.Printf("... не удалось\n %v\n", err)
		os.Exit(1) // Если не неврные логин и\или пароль, то выходим с кодом 1
	} else {
		fmt.Printf(".")
	}
	// Меняю папку на сервере
	if err := c.ChangeDir(Path); err != nil {
		fmt.Printf("... не удалось\n %v\n", err)
	} else {
		fmt.Printf(".  ОК\n")
	}
	return c
}

// DownloadFileFromFTP Download File From FTP with check file size file on FTP and local file
func DownloadFileFromFTP(c *ftp.ServerConn, Path, LocalPath, FileIn string) (codeExit int, err error) {
	//Узнаём размер файла
	sourceSize, _ := c.FileSize(Path + FileIn)
	//Проверяю возможность получания файла с FTP
	r, err := c.Retr(Path + FileIn)
	if err != nil {
		return 2, err
	}
	fmt.Printf("%v Скачиваю файл %s \n", time.Now().Format("15:04:05"), Path+FileIn)

	// Создаю файл для хранения
	dest, err := os.Create(LocalPath + FileIn + ".tmp")
	if err != nil {
		return 4, err
	}
	defer dest.Close()

	// Создаю прогрессбар
	bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.SetMaxWidth(80)
	bar.Start()

	// create proxy reader
	reader := bar.NewProxyReader(r)

	// and copy from reader
	io.Copy(dest, reader)
	bar.Finish()
	dest.Close() // Закрываю файл окрытый ранее
	r.Close()    // закрываю соединение
	return 0, nil
}

func checkDownloadedFile(c *ftp.ServerConn, Path, FileIn, LocalPath string) (int, error) {
	//-----------------------------------------
	/*Сравниваю размеры удаленного и локального файлов.
	Если они равны, то переименовываем временный файл
	Если они не равны, выдаём ошибку , удаляем временный файл
	*/
	// Получаю информацию о локальном файле
	fmt.Printf("%v Проверяю файл ", time.Now().Format("15:04:05"))
	diff, err := GetDiffFilesSize(c, Path+FileIn, LocalPath+FileIn+".tmp")
	if err != nil {
		fmt.Printf("не удалось: %v\n", err)
		return 6, err
	}
	if diff != 0 {
		fmt.Printf("%v \nФайлы не совпадают. Удаляю временный файл.", time.Now().Format("15:04:05"))
		if err := os.Remove(LocalPath + FileIn + ".tmp"); err != nil {
			fmt.Printf("Не удалось удалить %v : %v\n Удалите файл в ручную.", LocalPath+FileIn+".tmp", err)
			return 6, err
		}
		fmt.Printf("\t...\tОК\n")
		err := errors.New("Файлы не совпадают")
		return 6, err
	}
	fmt.Printf("\t...\tОК\n")
	if err := os.Rename(LocalPath+FileIn+".tmp", LocalPath+FileIn); err != nil {
		fmt.Printf("Не удалось переименовать файл %v: %v\n", LocalPath+FileIn+".tmp", err)
		return 6, err
	}
	return 0, nil
}

//UploadFileToFTP Upload File to FTP with check file size file on FTP and local file
func UploadFileToFTP(c *ftp.ServerConn, Path, LocalPath, FileOut string) (codeExit int, err error) {
	//Проверяю существование файла на локальном ПК
	if _, err := os.Stat(LocalPath + FileOut); os.IsNotExist(err) {
		fmt.Printf("%v Нет локально файла, нечего выгружать \n", time.Now().Format("15:04:05"))
		return 7, err
	}
	//Открываю файл для передачи на FTP
	fmt.Printf("%v Загружаю на сервер файл %v\n", time.Now().Format("15:04:05"), LocalPath+FileOut)
	source, err := os.Open(LocalPath + FileOut)
	defer source.Close()
	if err != nil {
		fmt.Printf("%v Не удалось открыть файл на локальном ПК, его не существует или он занят другой программой\n", time.Now().Format("15:04:05"))
		fmt.Printf("%v %v\n", time.Now().Format("15:04:05"), err)
		return 7, err
	}
	fmt.Printf("\t.")

	// get the size	//
	fi, err := os.Stat(LocalPath + FileOut)
	if err != nil {
		return 0, err
	}
	sourceSize := fi.Size()

	// create bar
	bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.SetMaxWidth(80)
	bar.Start()

	// create proxy reader
	reader := bar.NewProxyReader(source)

	//Заливаю файл на FTP
	if err = c.Stor(Path+FileOut, reader); err != nil {
		fmt.Printf("Не удалось:\n %v\n", err)
		return 2, err
	}

	// finish bar
	bar.Finish()
	return 0, nil
}

func checkUploadedFile(c *ftp.ServerConn, Path, FileOut, LocalPath string) (int, error) {
	//-----------------------------------------
	/*Сравниваю размеры удаленного и локального файлов.
	Если они равны, то переименовываем временный файл
	*/
	// Получаю информацию о локальном файле
	fmt.Printf("%v Проверяю файл ", time.Now().Format("15:04:05"))
	diff, err := GetDiffFilesSize(c, Path+FileOut, LocalPath+FileOut)
	if err != nil {
		fmt.Printf("не удалось: %v\n", err)
		return 6, err
	}
	if diff != 0 {
		fmt.Printf("%v \nФайлы не совпадают. Удаляю файл на FTP.", time.Now().Format("15:04:05"))
		if err := c.Delete(Path + FileOut); err != nil {
			fmt.Printf("Не удалось удалить %v . Удалите файл в ручную: %v\n", Path+FileOut, err)
			return 6, err
		}
	}
	return 0, nil
}

//FileOnFtpNotExist check exist file on ftp, and return xor value
func FileOnFtpNotExist(c *ftp.ServerConn, Path, FileIn string) bool {
	entries, err := c.NameList(Path)
	if err != nil {
		fmt.Printf("Не удалось проверить наличие файла на сервере %v", err)
		return true
	}

	for _, entry := range entries {
		if entry == Path+FileIn {
			return false
		}
	}

	return true
}

// GetDiffFilesSize Compare Size of two Files diff = sizeFullNameFileOnFtp - sizeFullNameLocalFile
func GetDiffFilesSize(c *ftp.ServerConn, fullNameFileOnFtp, fullNameLocalFile string) (diff int64, err error) {
	fi, err := os.Stat(fullNameLocalFile)
	if err != nil {
		return 0, err
	}
	// get the size
	sizeFullNameLocalFile := fi.Size()

	// Получаю размер удаленного файла
	sizeFullNameFileOnFtp, err := c.FileSize(fullNameFileOnFtp)
	if err != nil {
		return 0, err
	}
	diff = sizeFullNameFileOnFtp - sizeFullNameLocalFile
	return diff, nil
}

func checkErrorCode(codeExit int) {
	switch codeExit {
	case 1:
		fmt.Printf("%v Ошибка соединения\n", time.Now().Format("15:04:05"))
	case 2:
		fmt.Printf("%v Не удалось открыть файл на FTP, его не существует или он занят другой программой\n", time.Now().Format("15:04:05"))
	case 3:
		fmt.Printf("%v Не удалось скачать файл - Сетевая ошибка\n", time.Now().Format("15:04:05"))
	case 4:
		fmt.Printf("%v Не удалось открыть локальный файл на запись, возможно он занят другой программой\n", time.Now().Format("15:04:05"))
	case 5:
		fmt.Printf("%v Запись файла не возможна\n", time.Now().Format("15:04:05"))
	case 6:
		fmt.Printf("%v Ошибка проверки файла\n", time.Now().Format("15:04:05"))
	case 7:
		fmt.Printf("%v Не удалось открыть локальный файл, его не существутет или он занят другой программой\n", time.Now().Format("15:04:05"))
	case 8:
		fmt.Printf("%v Не известная ошибка\n", time.Now().Format("15:04:05"))
	case 9:
		fmt.Printf("%v Не известная ошибка\n", time.Now().Format("15:04:05"))
	}
}

func main() {
	intro()
	// Получаю конфиг из файла
	cfg.LoadConfig(configFileName)
	c := ConnectToFTP(cfg.ServerPort, cfg.Login, cfg.Password, cfg.Path)
	// заканчиваю, выходу и закрываю соединение
	defer c.Logout()
	defer c.Quit()

	fmt.Println(strings.Repeat("-", 80))

	if FileOnFtpNotExist(c, cfg.Path, cfg.FileIn) { //Если файла на FTP не существует то выдаём ошибку
		err := errors.New("На сервере нет файла, нечего загружать")
		fmt.Printf("%v %v\n", time.Now().Format("15:04:05"), err)
	} else {
		// codeExit, err = DownloadFileFromFTP(c, Path, LocalPath, FileIn)
		codeExit, _ = DownloadFileFromFTP(c, cfg.Path, cfg.LocalPath, cfg.FileIn)
		if codeExit == 0 {
			codeExit, err = checkDownloadedFile(c, cfg.Path, cfg.FileIn, cfg.LocalPath)
			if err != nil {
				checkErrorCode(codeExit)
			} else {
				fmt.Printf("%v Файл скачан, удаляю на сервере", time.Now().Format("15:04:05"))
				if err := c.Delete(cfg.Path + cfg.FileIn); err != nil {
					fmt.Printf("\t...\tНе получилось.\n")
				} else {
					fmt.Printf("\t...\tOK\n")
				}
			}
		}
	}

	codeExit, err = UploadFileToFTP(c, cfg.Path, cfg.LocalPath, cfg.FileOut)
	if codeExit == 0 {
		codeExit, err = checkUploadedFile(c, cfg.Path, cfg.FileOut, cfg.LocalPath)
		if err != nil {
			checkErrorCode(codeExit)
		} else {
			fmt.Printf("%v Файл закачан на сервер, удаляю с ПК", time.Now().Format("15:04:05"))
			if err := os.Remove(cfg.LocalPath + cfg.FileOut); err != nil {
				fmt.Printf("\t...\tНе получилось.\n")
			} else {
				fmt.Printf("\t...\tOK\n")
			}
		}
	}
	fmt.Println(strings.Repeat("=", 80))

	os.Exit(codeExit)
}
