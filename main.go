package main

import (
	"log"
	"os"
	"strconv"
)

//PortalAPIKey é a chave de acesso a api de osm do portal
var PortalAPIKey string

//PortalURL é o endereço da api de osm do portal
var PortalURL string

//CoswinURL é a url do SOAP do coswin fm2c
var CoswinURL string

//CoswinUsername é o nome de usuario do SOAP coswin fm2c
var CoswinUsername string

//CoswinPassword é a senha de usuario do SOAP coswin fm2c
var CoswinPassword string

//CoswinDatasource é o nome do datasource do SOAP coswin fm2c
var CoswinDatasource string

func main() {
	PortalAPIKey = os.Getenv("PORTAL_API_KEY")
	PortalURL = os.Getenv("PORTAL_URL")
	CoswinURL = os.Getenv("COSWIN_URL")
	CoswinUsername = os.Getenv("COSWIN_USERNAME")
	CoswinPassword = os.Getenv("COSWIN_PASSWORD")
	CoswinDatasource = os.Getenv("COSWIN_DATASOURCE")
	if PortalAPIKey == "" || PortalURL == "" || CoswinURL == "" || CoswinUsername == "" || CoswinPassword == "" || CoswinDatasource == "" {
		log.Fatal("variáveis de ambiente não definidos (PORTAL_API_KEY, PORTAL_URL, COSWIN_URL, COSWIN_USERNAME, COSWIN_PASSWORD, COSWIN_DATASOURCE)")
	}

	templateCreate, err := loadTemplate("./templates/create.xml")
	if err != nil {
		log.Fatal(err)
	}
	templateUpdate, err := loadTemplate("./templates/update.xml")
	if err != nil {
		log.Fatal(err)
	}
	templateFind, err := loadTemplate("./templates/find.xml")
	if err != nil {
		log.Fatal(err)
	}

	pacoteOSM, err := carregarOSM()
	if err != nil {
		log.Fatal(err)
	}

	coswinAcompanhamento := []int{}
	for _, osm := range pacoteOSM.Acompanhamento {
		wowocode, err := strconv.Atoi(osm.Coswin)
		if err == nil && wowocode > 0 {
			coswinAcompanhamento = append(coswinAcompanhamento, wowocode)
		}
	}
	minWowo, maxWowo := minmax(coswinAcompanhamento)
	if err := acompanhamentoCoswin(minWowo, maxWowo, pacoteOSM.Acompanhamento, templateFind); err != nil {
		log.Println(err)
	}

	for _, osm := range pacoteOSM.Reabertura {
		if err := reaberturaCoswin(osm, templateUpdate); err != nil {
			log.Println(err)
		}
	}

	for _, osm := range pacoteOSM.Novo {
		criarWoWo(osm, templateCreate, templateUpdate)
	}
}
