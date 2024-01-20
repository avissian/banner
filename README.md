# Запуск:

banner.exe tlo_config /var/www/webtlo/src/data/config.ini --ban_list=ban_list.txt
Параметры:
 * tlo_config - обязательно, после этого указать путь к файлу конфига webTLO
 * --ban_list=ban_list.txt - путь к файлу со списком клиентов, которых необходимо банить, ищется подстрокой, без учёта регистра