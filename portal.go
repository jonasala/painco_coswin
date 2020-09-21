package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type localmotivo struct {
	CodCoswin string `json:"cod_coswin"`
}

type osm struct {
	Codigo               int         `json:"codigo"`
	Solicitante          string      `json:"nome_solicitante"`
	Local                localmotivo `json:"local"`
	Motivo               localmotivo `json:"motivo"`
	DataPrevisao         string      `json:"data_previsao"`
	Descricao            string      `json:"descricao"`
	ObservacaoAnalise    string      `json:"observacao_analise"`
	ObservacaoReabertura string      `json:"observacao_reabertura"`
	Revisao              string      `json:"revisao"`
	Coswin               string      `json:"coswin"`
}

type pacote struct {
	Novo           []osm `json:"novo"`
	Acompanhamento []osm `json:"acompanhamento"`
}

func carregarOSM() (*pacote, error) {
	response, err := http.Get(
		PortalURL + "/listar_osm.php?API_KEY=" + PortalAPIKey,
	)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	osm := pacote{}
	if err := json.Unmarshal(body, &osm); err != nil {
		return nil, err
	}

	log.Printf("Carregar OSM:\n%+v\n", osm)

	return &osm, nil
}

func atualizarOSM(codigoOSM int, data map[string]interface{}) error {
	requestBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	response, err := http.Post(
		PortalURL+"/atualizar_osm.php?osm="+strconv.Itoa(codigoOSM)+"&API_KEY="+PortalAPIKey,
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	log.Printf("Atualizar OSM:---REQUEST---\n%v\n\n---RESPONSE---\n%v\n", string(requestBody), string(body))

	return nil
}
