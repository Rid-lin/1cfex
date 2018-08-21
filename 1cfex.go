// main project main.go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/jlaffaye/ftp"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var configExample = `#Пример конфигурационного файла
# Порядок не имеет значения
login_FTP = login
pass_FTP = password
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
1cfex 1.1.0                     Vlad Vegner                    August 21th 2018
===============================================================================
`)
}

//PrinT - Custom Printf with time
func PrinT(format string, args ...interface{}) {
	fmt.Printf(time.Now().Format("15:04:04") + " " + fmt.Sprintf(format, args...))
}

//PrintOK Print OK
func PrintOK() {
	fmt.Printf("\t...\tOK\n")
}

//GetConfig - Get config from commandline
func (s *ConfigAttr) GetConfig() {
	flag.StringVar(&s.ServerPort, "ServerPort", "10.57.254.103:21", "Сервер и порт к которому необходимо подключиться(например 10.57.254.103:21)")
	flag.StringVar(&s.Login, "Login", "kust", "Логин для входа на сервер")
	flag.StringVar(&s.Password, "Password", "", "Пароль для входа на сервер")
	flag.StringVar(&s.Path, "Path", "/srv/1cv8/uat/", "Папка из которой будет браться файл с сервера")
	flag.StringVar(&s.LocalPath, "LocalPath", "C:/ftpswap/LocalObmenUAT/", "Локальная папка из которой будет браться файл для выгрузки на сервер")
	flag.StringVar(&s.FileIn, "FileIn", "", "Файл для загрузки")
	flag.StringVar(&s.FileOut, "FileOut", "", "Файл для выгрузки")
	flag.Parse()
}

//LoadConfig - Load config from ini-file
func (s *ConfigAttr) LoadConfig(nameFile string) {
	PrinT("Загружаю настройки из %v", nameFile)
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
	PrinT("Устанавливаю соедиение с сервером %s\n", ServerPort)
	PrinT("Авторизуюсь, меняю папку \t")
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
	PrinT("Скачиваю файл %s \n", Path+FileIn)

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

//checkDownloadedFile check downloaded file from FTP
func checkDownloadedFile(c *ftp.ServerConn, Path, FileIn, LocalPath string) (int, error) {
	//-----------------------------------------
	/*Сравниваю размеры удаленного и локального файлов.
	Если они равны, то переименовываем временный файл
	Если они не равны, выдаём ошибку , удаляем временный файл
	*/
	// Получаю информацию о локальном файле
	PrinT("Проверяю файл ")
	diff, err := GetDiffFilesSize(c, Path+FileIn, LocalPath+FileIn+".tmp")
	if err != nil {
		fmt.Printf("... не удалось.\n %v\n", err)
		return 6, err
	}
	if diff != 0 {
		PrinT("\nФайлы не совпадают. Удаляю временный файл.")
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
		PrinT("Нет локально файла, нечего выгружать \n")
		return 7, err
	}
	//Открываю файл для передачи на FTP
	PrinT("Загружаю на сервер файл %v\n", LocalPath+FileOut)
	source, err := os.Open(LocalPath + FileOut)
	defer source.Close()
	if err != nil {
		PrinT("Не удалось открыть файл на локальном ПК, его не существует или он занят другой программой\n")
		PrinT("%v\n", err)
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

//checkUploadedFile check uploaded file on FTP
func checkUploadedFile(c *ftp.ServerConn, Path, FileOut, LocalPath string) (int, error) {
	//-----------------------------------------
	/*Сравниваю размеры удаленного и локального файлов.
	Если они равны, то переименовываем временный файл
	*/
	// Получаю информацию о локальном файле
	PrinT("Проверяю файл")
	diff, err := GetDiffFilesSize(c, Path+FileOut, LocalPath+FileOut)
	if err != nil {
		fmt.Printf("... не удалось.\n %v\n", err)
		return 6, err
	}
	if diff != 0 {
		PrinT("\nФайлы не совпадают. Удаляю файл на FTP.")
		if err := c.Delete(Path + FileOut); err != nil {
			fmt.Printf("Не удалось удалить %v . Удалите файл в ручную: %v\n", Path+FileOut, err)
			return 6, err
		}
	}
	PrintOK()

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

//checkErrorCode - get Error code and print error massage
func checkErrorCode(codeExit int) {
	switch codeExit {
	case 1:
		PrinT("Ошибка соединения\n")
	case 2:
		PrinT("Не удалось открыть файл на FTP, его не существует или он занят другой программой\n")
	case 3:
		PrinT("Не удалось скачать файл - Сетевая ошибка\n")
	case 4:
		PrinT("Не удалось открыть локальный файл на запись, возможно он занят другой программой\n")
	case 5:
		PrinT("Запись файла не возможна\n")
	case 6:
		PrinT("Ошибка проверки файла\n")
	case 7:
		PrinT("Не удалось открыть локальный файл, его не существутет или он занят другой программой\n")
	case 8:
		PrinT("Не известная ошибка\n")
	case 9:
		PrinT("Не известная ошибка\n")
	}
}

func main() {
	intro()
	// Смотрю есть ли данные в командной строке
	cfg.GetConfig()
	// Если в аргументах командной строки нет входящего файла, то пытаемся прочитать конфигурационный файл
	if cfg.FileIn == "" {
		// Получаю конфиг из файла
		cfg.LoadConfig(configFileName)
	}
	c := ConnectToFTP(cfg.ServerPort, cfg.Login, cfg.Password, cfg.Path)
	// заканчиваю, выхожу и закрываю соединение
	defer c.Logout()
	defer c.Quit()

	fmt.Println(strings.Repeat("-", 80))

	if FileOnFtpNotExist(c, cfg.Path, cfg.FileIn) { //Если файла на FTP не существует то выдаём ошибку
		err := errors.New("На сервере нет файла, нечего загружать")
		PrinT("%v\n", err)
		// fmt.Printf("%v %v\n", time.Now().Format("15:04:05"), err)
	} else {
		codeExit, _ = DownloadFileFromFTP(c, cfg.Path, cfg.LocalPath, cfg.FileIn)
		if codeExit == 0 {
			codeExit, err = checkDownloadedFile(c, cfg.Path, cfg.FileIn, cfg.LocalPath)
			if err != nil {
				checkErrorCode(codeExit)
			} else {
				PrinT("Файл скачан, удаляю на сервере")
				if err := c.Delete(cfg.Path + cfg.FileIn); err != nil {
					fmt.Printf("\t...\tНе получилось.\n")
				} else {
					PrintOK()
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
			PrinT("Файл закачан на сервер, удаляю с ПК")
			if err := os.Remove(cfg.LocalPath + cfg.FileOut); err != nil {
				fmt.Printf("\t...\tНе получилось.\n")
			} else {
				PrintOK()
			}
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	PrinT("Отключаюсь от сервера")
	// заканчиваю, выхожу и закрываю соединение
	c.Logout()
	c.Quit()
	PrintOK()
	fmt.Println(strings.Repeat("=", 80))

	os.Exit(codeExit)
}
