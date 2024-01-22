package tlo_config

import (
	"gopkg.in/ini.v1"
	"regexp"
	"strconv"
)

// ConfigT base type
type ConfigT struct {
	Clients    []ClientT
	Proxy      ProxyT
	Tracker    TrackerT
	CatList    []string
	Categories []CategoryT
	path       string
	source     *ini.File
}

/** subtypes **/

type ClientT struct {
	Type  string //transmission,qbittorrent...
	Name  string
	Login string
	Pass  string
	Host  string
	Port  uint32
	SSL   bool
}

type ProxyT struct {
	ActivateForum bool
	ActivateApi   bool
	Type          string
	Host          string
	Port          uint32
	Login         string
	Pass          string
}

type TrackerT struct {
	ApiURL      string
	ForumURL    string
	Login       string
	Pass        string
	UserID      uint64
	BTKey       string
	APIKey      string
	APISsl      bool
	ForumSsl    bool
	UserSession string
}

type CategoryT struct {
	Num string
	//Title string
	//client          = 1
	Label      string
	DataFolder string
	//DataSubFolder = 1
	//hide-topics     = 0
	//control-peers   =
	//exclude         = 0
}

/** methods **/

func (config *ConfigT) Load(path string) (err error) {
	config.path = path

	cfg, err := ini.Load(path)
	if err != nil {
		return
	}
	config.source = cfg

	config.Clients, err = loadClients(cfg)
	if err != nil {
		return
	}
	config.Proxy = loadProxy(cfg)
	if err != nil {
		return
	}
	config.Tracker = loadTracker(cfg)
	config.Categories = loadCategories(cfg)
	if err != nil {
		return
	}

	if cl, err := cfg.Section("sections").GetKey("subsections"); err == nil {
		config.CatList = cl.Strings(",")
	}

	return

}

func loadClients(cfg *ini.File) (clients []ClientT, err error) {
	if qt, err := cfg.Section("other").GetKey("qt"); err == nil {
		if qtNum, err := qt.Int(); err == nil {
			for clientNum := 1; clientNum <= qtNum; clientNum++ {
				var client ClientT
				clientCfg := cfg.Section("torrent-client-" + strconv.Itoa(clientNum))
				if typeClient, err := clientCfg.GetKey("client"); err == nil {
					client.Type = typeClient.String()
				}
				if name, err := clientCfg.GetKey("comment"); err == nil {
					client.Name = name.String()
				}
				if login, err := clientCfg.GetKey("login"); err == nil {
					client.Login = login.String()
				}
				if pass, err := clientCfg.GetKey("password"); err == nil {
					client.Pass = pass.String()
				}
				if host, err := clientCfg.GetKey("hostname"); err == nil {
					client.Host = host.String()
				}
				if port, err := clientCfg.GetKey("port"); err == nil {
					if portI, err := port.Uint64(); err == nil {
						client.Port = uint32(portI)
					}
				}
				if ssl, _ := clientCfg.GetKey("ssl"); err == nil {
					client.SSL, _ = ssl.Bool()
				}
				clients = append(clients, client)
			}
		}
	}
	return
}

func loadProxy(cfg *ini.File) (proxy ProxyT) {
	proxyCfg := cfg.Section("proxy")

	if activateForum, err := proxyCfg.GetKey("activate_forum"); err == nil {
		proxy.ActivateForum, _ = activateForum.Bool()
	}
	if activateApi, err := proxyCfg.GetKey("activate_api"); err == nil {
		proxy.ActivateApi, _ = activateApi.Bool()
	}
	if typeS, err := proxyCfg.GetKey("type"); err == nil {
		proxy.Type = typeS.String()
	}
	if host, err := proxyCfg.GetKey("hostname"); err == nil {
		proxy.Host = host.String()
	}
	if portS, err := proxyCfg.GetKey("port"); err == nil {
		portI, _ := portS.Uint64()
		proxy.Port = uint32(portI)
	}
	if login, err := proxyCfg.GetKey("login"); err == nil {
		proxy.Login = login.String()
	}
	if pass, err := proxyCfg.GetKey("password"); err == nil {
		proxy.Pass = pass.String()
	}

	return
}
func loadTracker(cfg *ini.File) (tracker TrackerT) {
	trackerCfg := cfg.Section("torrent-tracker")

	if apiURL, err := trackerCfg.GetKey("api_url"); err == nil {
		tracker.ApiURL = apiURL.String()
		if apiURL, err = trackerCfg.GetKey("api_url_custom"); tracker.ApiURL == "custom" && err == nil {
			tracker.ApiURL = apiURL.String()
		}
	}
	if forumURL, err := trackerCfg.GetKey("forum_url"); err == nil {
		tracker.ForumURL = forumURL.String()
		if forumURL, err = trackerCfg.GetKey("forum_url_custom"); tracker.ForumURL == "custom" && err == nil {
			tracker.ForumURL = forumURL.String()
		}
	}
	if login, err := trackerCfg.GetKey("login"); err == nil {
		tracker.Login = login.String()
	}
	if pass, err := trackerCfg.GetKey("password"); err == nil {
		tracker.Pass = pass.String()
	}
	if userID, err := trackerCfg.GetKey("user_id"); err == nil {
		tracker.UserID, _ = userID.Uint64()

	}
	if btKEY, err := trackerCfg.GetKey("bt_key"); err == nil {
		tracker.BTKey = btKEY.String()

	}
	if apiKEY, err := trackerCfg.GetKey("api_key"); err == nil {
		tracker.APIKey = apiKEY.String()

	}
	if apiSSL, err := trackerCfg.GetKey("api_ssl"); err == nil {
		tracker.APISsl, _ = apiSSL.Bool()

	}
	if forumSSL, err := trackerCfg.GetKey("forum_ssl"); err == nil {
		tracker.ForumSsl, _ = forumSSL.Bool()
	}
	if userSession, err := trackerCfg.GetKey("user_session"); err == nil {
		tracker.UserSession = userSession.String()
	}
	return
}

func loadCategories(cfg *ini.File) (categories []CategoryT) {
	re := regexp.MustCompile("^\\d+$")
	for _, sec := range cfg.Sections() {

		if re.MatchString(sec.Name()) {
			/*title, err := sec.GetKey("title")
			if err != nil {
				return nil, err
			}*/
			if label, err := sec.GetKey("label"); err == nil {
				if df, err := sec.GetKey("data-folder"); err == nil {
					var c CategoryT
					c.Num = sec.Name()
					//c.Title = title.String()
					c.Label = label.String()
					c.DataFolder = df.String()

					categories = append(categories, c)
				}
			}
		}

	}
	return
}

// TODO: ломает ТЛО из-за кавычек
//func (config *ConfigT) Save() (err error) {
//	for _, val := range config.Categories {
//		k, _ := config.source.Section(val.Num).GetKey("data-folder")
//		k.SetValue(val.DataFolder)
//	}
//	err = config.source.SaveTo(config.path)
//	return
//}
