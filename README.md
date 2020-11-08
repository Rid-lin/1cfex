# 1cfex

Простая программа для периодического автоматического обмена фалйами через FTP (разработана для осуществления обмена файлами базы данных 1С)

Коды завершения программы:

1. Ошибка соединения
2. Не удалось открыть файл на FTP, его не существует или он занят другой программой  
3. Не удалось скачать файл - Сетевая ошибка  
4. Не удалось открыть локальный файл на запись, возможно он занят другой программой
5. Запись файла не возможна
6. Ошибка проверки файла - Это когда скаченный/закаченный файл по объёму не совпадает с оригиналом
7. Не удалось открыть локальный файл, его не существует или он занят другой программой
8. Не известная ошибка - не придумал названия для нее
9. Не известная ошибка - не придумал названия для нее

Параметры командной строки можно просмотреть набрав в консоли 1cfex.exe -h

- *-FileIn*  - Файл для загрузки на сервер
- *-FileOut*  - Файл для выгрузки с сервера
- *-LocalPath* - Локальная папка из которой будет браться файл для выгрузки на сервер (по-умолчанию "C:/ftpswap/LocalObmenUAT/")
- *-Login* - Логин для входа на сервер (по-умолчанию "kust")
- *-Password* - Пароль для входа на сервер
- *-Path* - Папка из которой будет браться файл с сервера (по-умолчанию "/srv/1cv8/uat/")
- *-ServerPort* - Сервер и порт к которому необходимо подключиться(например 10.57.254.103:21) (default "10.57.254.103:21")

Программа может также запускаться и с конфигурационным файлом 1cfex.ini , параметры совпадают с командной строкой.

Вот его содержание

    login_FTP =
    pass_FTP =
    path =
    local_path =
    server =
    file_in =
    file_out =


A simple program for periodic automatic exchange of files via FTP (developed for the exchange of 1C database files)

Program completion codes:

1. Connection error
2. Failed to open file on FTP, it does not exist or it is busy with another program
3. Failed to download file - Network error
4. Failed to open a local file for writing, it may be busy with another program
5. File writing is not possible
6. File verification error - This is when the downloaded / uploaded file does not match the original in size
7. The local file could not be opened, it does not exist or it is busy with another program
8. Unknown error - didn't come up with a name for it
9. Unknown error - did not come up with a name for it

Command line parameters can be viewed by typing 1cfex.exe -h in the console

- *-FileIn* - File to upload to the server
- *-FileOut* - File to upload from the server
- *-LocalPath* - Local folder from which the file will be taken for uploading to the server (by default "C: / ftpswap / LocalObmenUAT /")
- *-Login* - Login to enter the server (by default "kust")
- *-Password* - Password for entering the server
- *-Path* - The folder from which the file will be taken from the server (by default "/ srv / 1cv8 / uat /")
- *-ServerPort* - Server and port to which you need to connect (for example 10.57.254.103:21) (default "10.57.254.103:21")

The program can also be launched with the 1cfex.ini configuration file, the parameters are the same as the command line.

Here is its content

    login_FTP =
    pass_FTP =
    path =
    local_path =
    server =
    file_in =
    file_out =
