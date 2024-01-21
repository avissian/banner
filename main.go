package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/avissian/banner/tlo_config"
	"github.com/avissian/go-qbittorrent/qbt"
	"log"
	"os"
	"strconv"
	"strings"
)

type LiteTorrent struct {
	Magnet   string
	SavePath string
	Name     string
}

func Connect(user string, pass string, server string, port uint32, SSL bool) (client *qbt.Client) {
	scheme := "http"
	if SSL {
		scheme = "https"
	}
	lo := qbt.LoginOptions{Username: user, Password: pass}
	client = qbt.NewClient(fmt.Sprintf("%s://%s:%d", scheme, server, port))
	if err := client.Login(lo); err != nil {
		log.Panicln(err)
	}
	//ver, err := client.BuildInfo()
	//if err != nil {
	//		log.Panicln(err)
	//}
	//log.Printf("Connected (%s:%d): %#v", server, port, ver)

	return
}

func getStrangePeers(login string, pass string, host string, port uint32, ssl bool, banPath string) (IP2ban []string) {
	var clients []string

	file, err := os.Open(banPath)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		clients = append(clients, strings.ToLower(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	client := Connect(login, pass, host, port, ssl)
	s := "active"
	torrentList, err := client.Torrents(qbt.TorrentsOptions{Filter: &s})
	if err != nil {
		log.Panicln(err)
	}

	if *verbose {
		log.Printf("Active torrents: %d", len(torrentList))
	}
	for _, torrent := range torrentList {
		peers, err := client.TorrentPeers(torrent.Hash, 0)
		if len(peers.Peers) == 0 {
			continue
		}
		if err != nil {
			continue
		}
		if *verbose {
			var names []string
			for ipPort, peer := range peers.Peers {
				if strings.TrimSpace(peer.Client) == "" {
					peer.Client = ipPort
				}
				names = append(names, peer.Client)
			}

			log.Printf("%3d peers: %v", len(peers.Peers), "[ "+strings.Join(names, ", ")+" ]")
		}
		for ipPort, peer := range peers.Peers {
			if torrent.TotalSize > 0 && (*trafLimit) > 0 && peer.Uploaded > int64(float64(torrent.TotalSize)*(*trafLimit)) {
				log.Printf("Blocked by traffic: %#v %d", peer, torrent.TotalSize)
				IP2ban = append(IP2ban, ipPort)
			}
			for _, client := range clients {

				if strings.Contains(strings.ToLower(peer.Client), client) {
					log.Printf("Blocked by client name: %#v %s", peer, torrent.Name)
					IP2ban = append(IP2ban, ipPort)
				}
			}
		}
	}
	return
}

func ban(login string, pass string, host string, port uint32, ssl bool, ip2ban []string) {
	if len(ip2ban) > 0 {
		client := Connect(login, pass, host, port, ssl)
		err := client.BanPeers(ip2ban)
		if err != nil {
			log.Printf("ERROR: %#v", err)
		}
		log.Printf("Banned: %v", ip2ban)
	}
}

func tloConfig(path string, banPath string) {
	var tlo tlo_config.ConfigT
	err := tlo.Load(path)
	if err != nil {
		log.Panicln(err)
	}
	if *verbose {
		log.Printf("Clients: %d", len(tlo.Clients))
	}
	var ip2ban []string
	for _, clientCfg := range tlo.Clients {
		ip2ban = append(ip2ban, getStrangePeers(clientCfg.Login, clientCfg.Pass, clientCfg.Host, clientCfg.Port, clientCfg.SSL, banPath)...)
	}
	for _, clientCfg := range tlo.Clients {
		ban(clientCfg.Login, clientCfg.Pass, clientCfg.Host, clientCfg.Port, clientCfg.SSL, ip2ban)
	}
}

var (
	//
	tloConfigCommand = kingpin.Command("tlo_config", "Путь к config.ini WebTLO (для тех кто в теме), заменяет все другие параметры.")
	configPath       = tloConfigCommand.Arg("path", "Путь к файлу конфига").Required().File()

	//
	clientParamsCommand = kingpin.Command("qbittorrent", "Manual qBittorrent params. Type \""+os.Args[0]+" qbittorrent --help\" for details.")
	server              = clientParamsCommand.Arg("server", "server:port").Required().String()
	user                = clientParamsCommand.Flag("user", "qBittorrent user.").Short('u').String()
	pass                = clientParamsCommand.Flag("pass", "qBittorrent password.").Short('p').String()
	SSL                 = clientParamsCommand.Flag("ssl", "SSL/TLS (https), default: no SSL/TLS (http)").Bool()
	//
	banList   = kingpin.Flag("ban_list", "Путь к файлу списка подстрок для бана клиентов").Short('b').Required().File()
	trafLimit = kingpin.Flag("traff", "Коэффициент превышения скачивания раздачи над её размером для бана пира").Default("1.1").Float64()
	verbose   = kingpin.Flag("verbose", "Подробный вывод лога").Short('v').Bool()
)

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}
	command := kingpin.Parse()
	if *verbose {
		printVersion()
	}
	switch command {
	case "tlo_config":
		tloConfig((*configPath).Name(), (*banList).Name())
	case "qbittorrent":
		port := uint32(8080)
		serverParams := strings.Split(*server, ":")
		host := serverParams[0]
		if len(serverParams) > 1 {
			portI, _ := strconv.Atoi(serverParams[1])
			port = uint32(portI)
		}
		ban(*user, *pass, host, port, *SSL,
			getStrangePeers(*user, *pass, host, port, *SSL, (*banList).Name()))
	default:
		log.Panicln("Unknown command")
	}
}
