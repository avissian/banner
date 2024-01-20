package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/avissian/go-qbittorrent/qbt"
	"github.com/ungerik/go-dry"
	"log"
	"os"
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

func getStrangePeers(path string, banPath string, debugLog *bool) {
	var tlo ConfigT
	err := tlo.Load(path)
	dry.PanicIfErr(err)
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

	var banIP []string
	if *debugLog {
		log.Printf("Clients: %d", len(tlo.Clients))
	}
	for _, clientCfg := range tlo.Clients {
		client := Connect(clientCfg.Login, clientCfg.Pass, clientCfg.Host, clientCfg.Port, clientCfg.SSL)
		s := "active"
		torerntList, err := client.Torrents(qbt.TorrentsOptions{Filter: &s})
		if err != nil {
			continue
		}

		if *debugLog {
			log.Printf("Active torrents: %d", len(torerntList))
		}
		for idx, torrent := range torerntList {
			peers, err := client.TorrentPeers(torrent.Hash, 0)
			if err != nil {
				continue
			}
			if *debugLog {
				var names []string
				for _, peer := range peers.Peers {
					names = append(names, peer.Client)
				}

				log.Printf("torrent %3d peers: %v", idx, strings.Join(names, ", "))
			}
			for ipPort, peer := range peers.Peers {
				if peer.Uploaded > int64(float64(torrent.TotalSize)*1.1) {
					log.Printf("Blocked by traffic: %#v %d", peer, torrent.TotalSize)
					banIP = append(banIP, ipPort)
				}
				for _, client := range clients {

					if strings.Contains(strings.ToLower(peer.Client), client) {
						log.Printf("Blocked by client name: %#v %s", peer, torrent.Name)
						banIP = append(banIP, ipPort)
					}
				}
			}
		}
		err = client.BanPeers(banIP)
		if err != nil {
			log.Printf("ERROR: %#v", err)
		}
	}

}

var (
	tloConfig  = kingpin.Command("tlo_config", "Путь к config.ini WebTLO (для тех кто в теме), заменяет все другие параметры.")
	configPath = tloConfig.Arg("path", "Путь к файлу конфига").Required().File()
	//ban_list    = tloConfig.Arg("ban_list", "Путь к файлу списка подстрок для бана клиентов").Required().File()
	//clientParams = kingpin.Command("qbittorrent", "Manual qBittorrent params. Type \""+os.Args[0]+" qbittorrent --help\" for details.")
	//server        = clientParams.Arg("server", "server:port").Required().String()
	//user          = clientParams.Flag("user", "qBittorrent user.").String()
	//pass          = clientParams.Flag("pass", "qBittorrent password.").String()
	//SSL           = clientParams.Flag("ssl", "SSL/TLS (https), default: no SSL/TLS (http)").Bool()
	ban_list = kingpin.Flag("ban_list", "Путь к файлу списка подстрок для бана клиентов").Required().File()
	verbose  = kingpin.Flag("verbose", "Подробный вывод лога").Short('v').Bool()
)

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}

	switch kingpin.Parse() {
	case "tlo_config":
		if configPath == nil {
			log.Println("Nil =(")
		}
		getStrangePeers((*configPath).Name(), (*ban_list).Name(), verbose)
	default:
		log.Panicln("Unknown command")
	}
}
