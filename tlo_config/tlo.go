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
	config.Proxy, err = loadProxy(cfg)
	if err != nil {
		return
	}
	config.Tracker, err = loadTracker(cfg)
	if err != nil {
		return
	}
	config.Categories, err = loadCategories(cfg)
	if err != nil {
		return
	}

	cl, err := cfg.Section("sections").GetKey("subsections")
	if err != nil {
		return
	}
	config.CatList = cl.Strings(",")

	return

}
func (config *ConfigT) Save() (err error) {
	for _, val := range config.Categories {
		k, _ := config.source.Section(val.Num).GetKey("data-folder")
		k.SetValue(val.DataFolder)
	}
	err = config.source.SaveTo(config.path)
	return
}

func loadClients(cfg *ini.File) (clients []ClientT, err error) {
	qt, err := cfg.Section("other").GetKey("qt")
	if err != nil {
		return
	}
	qtNum, err := qt.Int()
	if err != nil {
		return
	}
	for clientNum := 1; clientNum <= qtNum; clientNum++ {
		clientCfg := cfg.Section("torrent-client-" + strconv.Itoa(clientNum))
		typeClient, _ := clientCfg.GetKey("client")
		name, _ := clientCfg.GetKey("comment")
		login, _ := clientCfg.GetKey("login")
		pass, _ := clientCfg.GetKey("password")
		host, _ := clientCfg.GetKey("hostname")
		port, _ := clientCfg.GetKey("port")
		portI, _ := port.Uint64()
		ssl, _ := clientCfg.GetKey("ssl")
		sslB, _ := ssl.Bool()
		client := ClientT{
			typeClient.String(),
			name.String(),
			login.String(),
			pass.String(),
			host.String(),
			uint32(portI),
			sslB}
		clients = append(clients, client)
	}
	return
}

func loadProxy(cfg *ini.File) (proxy ProxyT, err error) {
	proxyCfg := cfg.Section("proxy")
	activateForum, _ := proxyCfg.GetKey("activate_forum")
	activateForumB, _ := activateForum.Bool()
	activateApi, _ := proxyCfg.GetKey("activate_api")
	activateApiB, _ := activateApi.Bool()
	typeS, _ := proxyCfg.GetKey("type")
	host, _ := proxyCfg.GetKey("hostname")
	portS, _ := proxyCfg.GetKey("port")
	portI, _ := portS.Uint64()
	login, _ := proxyCfg.GetKey("login")
	pass, _ := proxyCfg.GetKey("password")

	proxy.ActivateForum = activateForumB
	proxy.ActivateApi = activateApiB
	proxy.Type = typeS.String()
	proxy.Host = host.String()
	proxy.Port = uint32(portI)
	proxy.Login = login.String()
	proxy.Pass = pass.String()
	return
}
func loadTracker(cfg *ini.File) (tracker TrackerT, err error) {
	trackerCfg := cfg.Section("torrent-tracker")
	apiURL, _ := trackerCfg.GetKey("api_url")
	//apiURLCust, _ := trackerCfg.GetKey("api_url_custom")
	forumURL, _ := trackerCfg.GetKey("forum_url")
	login, _ := trackerCfg.GetKey("login")
	pass, _ := trackerCfg.GetKey("password")
	userID, _ := trackerCfg.GetKey("user_id")
	btKEY, _ := trackerCfg.GetKey("bt_key")
	apiKEY, _ := trackerCfg.GetKey("api_key")
	apiSSL, _ := trackerCfg.GetKey("api_ssl")
	forumSSL, _ := trackerCfg.GetKey("forum_ssl")
	userSession, _ := trackerCfg.GetKey("user_session")

	tracker.ApiURL = apiURL.String()
	tracker.ForumURL = forumURL.String()
	tracker.Login = login.String()
	tracker.Pass = pass.String()
	tracker.UserID, _ = userID.Uint64()
	tracker.BTKey = btKEY.String()
	tracker.APIKey = apiKEY.String()
	tracker.APISsl, _ = apiSSL.Bool()
	tracker.ForumSsl, _ = forumSSL.Bool()
	tracker.UserSession = userSession.String()
	return
}

func loadCategories(cfg *ini.File) (categories []CategoryT, err error) {
	re := regexp.MustCompile("^\\d+$")
	for _, sec := range cfg.Sections() {

		if re.MatchString(sec.Name()) {
			/*title, err := sec.GetKey("title")
			if err != nil {
				return nil, err
			}*/
			label, err := sec.GetKey("label")
			if err != nil {
				return nil, err
			}
			df, err := sec.GetKey("data-folder")
			if err != nil {
				return nil, err
			}
			var c CategoryT
			c.Num = sec.Name()
			//c.Title = title.String()
			c.Label = label.String()
			c.DataFolder = df.String()

			categories = append(categories, c)
		}

	}
	return
}
