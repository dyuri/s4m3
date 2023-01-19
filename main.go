package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v2"
)

type Config struct {
	TileTypes []TileType `yaml:"tile_types"`
	Seed      int64      `yaml:"seed"`
	Width     int        `yaml:"width"`
	Height    int        `yaml:"height"`
	Addr      string     `yaml:"addr"`
}

var defaultConfig = Config{
	TileTypes: []TileType{
		{
			Sign:  "*",
			Color: "#ffff00",
		},
		{
			Sign:  "X",
			Color: "#88ff00",
		},
		{
			Sign:  "O",
			Color: "#0088ff",
		},
	},
	Seed:   0,
	Width:  20,
	Height: 20,
}

type model struct {
	config   Config
	board    Board
	gameOver bool
	cx       int
	cy       int
	points   int
}

var _model model
var modelMu sync.RWMutex

func setWebModel(m model) {
	modelMu.Lock()
	defer modelMu.Unlock()
	_model = m
}

func getWebModel() model {
	modelMu.RLock()
	defer modelMu.RUnlock()
	return _model
}

func (m model) Init() tea.Cmd {
	setWebModel(m)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ":
			m.points += m.board.Hit(m.cx, m.cy)
		case "up":
			m.cy--
		case "down":
			m.cy++
		case "left":
			m.cx--
		case "right":
			m.cx++
		}
	case tea.MouseMsg:
		fmt.Println(msg)
	}

	if m.cx < 0 {
		m.cx = 0
	}
	if m.cx >= m.config.Width {
		m.cx = m.config.Width - 1
	}
	if m.cy < 0 {
		m.cy = 0
	}
	if m.cy >= m.config.Height {
		m.cy = m.config.Height - 1
	}

	return m, nil
}

func (m model) View() string {
	return m.board.WithCursor(m.cx, m.cy) + fmt.Sprintf("Points: %d", m.points)
}

func initialModel(cfg Config) model {
	return model{
		config: cfg,
		board:  generateBoard(cfg),
	}
}

func saveConfig(path string, cfg Config) error {
	ycfg, err := yaml.Marshal(cfg)

	if err != nil {
		log.Println(err)
		return err
	}

	err2 := ioutil.WriteFile(path, ycfg, 0644)

	if err2 != nil {
		log.Println(err2)
		return err2
	}

	log.Println("Configuration saved to", path)
	return nil
}

func readConfig(path string) (Config, error) {
	var cfg Config

	if path == "" {
		path = "config.yaml"
	}

	ycfg, err := ioutil.ReadFile(path)

	if err != nil {
		// error reading config file, use default config and try to save it
		errS := saveConfig(path, defaultConfig)
		return defaultConfig, errS
	}

	err2 := yaml.Unmarshal(ycfg, &cfg)

	if err2 != nil {
		return defaultConfig, err2
	}

	return cfg, nil
}

func main() {
	configPath := flag.String("c", "config.yaml", "path to config file")

	cfg, err := readConfig(*configPath)

	if err != nil {
		log.Println(err)
	}

	if cfg.Addr != "" {
		go initApi(cfg)
	}

	p := tea.NewProgram(initialModel(cfg))
	if err := p.Start(); err != nil {
		fmt.Printf("could not start game: %v", err)
		os.Exit(1)
	}

}
