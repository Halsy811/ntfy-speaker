
Настройка планировщика:
	Запуск программы: wscript.exe
	Ключи:
		//B //Nologo "\\main.hydrogas.ru\NETLOGON\ADUserData\ntfy_speaker\starter.vbs" -s http://ntfy.main.hydrogas.ru -p 80 -t test-topik
		//B //Nologo "\\main.hydrogas.ru\NETLOGON\ADUserData\ntfy_speaker\self.host.vbs" -s http://ntfy2.main.hydrogas.ru -p 80

wscript.exe:
	//B - Пакетный режим: подавляются отображение ошибок и запросов сценария
	//Nologo - Не отображать сведения о программе во время выполнения

Starter.vbs - Запуск топика
self.host.vbs - Запуск топика с именем компьютера (переводит символы в нижний регистр!)

