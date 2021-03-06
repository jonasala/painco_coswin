package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func criarWoWo(osm osm, templateCreate, templateUpdate *template.Template) error {
	/* CRIAR WOWO*/
	dataPrevisao := osm.DataPrevisao
	if dataPrevisao == "" {
		dataPrevisao = time.Now().Format("2006-01-02")
	}

	descricao := strings.ReplaceAll(osm.Descricao, "<BR />", " ")
	descricao = strings.ReplaceAll(descricao, "  ", " ")
	if len(descricao) > 100 {
		descricao = descricao[0:100]
	}

	requestBody, err := executeTemplate(templateCreate, map[string]interface{}{
		"WowoScheduleDate":   dataPrevisao,
		"WowoEquipment":      osm.Local.CodCoswin,
		"WowoJobClass":       osm.Motivo.CodCoswin,
		"WowoJobDescription": descricao,
	})
	if err != nil {
		return err
	}

	response, err := http.Post(
		CoswinURL+"/services/workorder/createsimple",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	envCreate := envelopeCreate{}
	if err := xml.Unmarshal(body, &envCreate); err != nil {
		return err
	}
	/* CRIAR WOWO */

	/* ATUALIZAR WOWO */
	solicitante := osm.Solicitante
	if len(solicitante) > 20 {
		solicitante = solicitante[0:20]
	}

	requestBody, err = executeTemplate(templateUpdate, map[string]interface{}{
		"WowoCode":        envCreate.Body.WorkOrderCreateSimpleKey.WowoCode,
		"WowoReporter":    solicitante,
		"WowoString12":    fmt.Sprintf("%v.%v", osm.Codigo, osm.Revisao),
		"WowoJobActivity": "<![CDATA[---OBSERVAÇÕES DE ANÁLISE---<br/>" + osm.ObservacaoAnalise + "]]>",
	})
	if err != nil {
		return err
	}

	response, err = http.Post(
		CoswinURL+"/services/workorder/update",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()
	/* ATUALIZAR WOWO */

	atualizarOSM(osm.Codigo, map[string]interface{}{
		"osm_coswin": envCreate.Body.WorkOrderCreateSimpleKey.WowoCode,
	})

	return nil
}

func acompanhamentoCoswin(wowoMin, wowoMax int, listaOSM []osm, templateFind, templateUpdate *template.Template) error {
	/* Listar WOWO do Coswin */
	requestBody, err := executeTemplate(templateFind, map[string]interface{}{
		"min": wowoMin,
		"max": wowoMax,
	})
	if err != nil {
		return err
	}

	log.Printf("ACOMPANHAMENTO:\n----REQUEST----\n%v\n", requestBody.String())

	response, err := http.Post(
		CoswinURL+"/services/workorder/find/find",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Printf("Acomanhamento FIND:\n----RESPONSE----\n%v\n", string(body))

	envFind := envelopeFind{}
	if err := xml.Unmarshal(body, &envFind); err != nil {
		return err
	}
	for _, wowo := range envFind.Body.RetworkorderfindfindParameters.WorkorderfindList.Workorderfind {
		if wowo.WowoString12 == "" || wowo.WowoUserStatus != "N" {
			continue
		}

		infoOSM := strings.Split(wowo.WowoString12, ".")
		if len(infoOSM) != 2 {
			continue
		}

		var osmAlvo *osm
		for _, osm := range listaOSM {
			if strconv.Itoa(osm.Codigo) == infoOSM[0] {
				osmAlvo = &osm
				break
			}
		}
		if osmAlvo == nil {
			continue
		}

		log.Printf("Acompanhamento:\n----ALVO:----\n%+v", osmAlvo)

		if osmAlvo.Revisao != infoOSM[1] {
			log.Println("AÇÃO: REABERTURA")
			return reaberturaCoswin(*osmAlvo, templateUpdate)
		}

		log.Println("AÇÃO: RESOLUÇÃO")
		dataFim, err := time.Parse("2006-01-02T15:04:05", wowo.WowoEndDate)
		if err != nil {
			dataFim = time.Now()
		}

		atualizarOSM(osmAlvo.Codigo, map[string]interface{}{
			"osm_status":           "R",
			"osm_observacao_final": strings.ReplaceAll(wowo.WowoFeedbackNote, "<br>", "\n"),
			"osm_data_hora_fim":    dataFim.Format("2006-01-02 15:04:05"),
		})
	}
	/* Listar WOWO do Coswin */
	return nil
}

func reaberturaCoswin(osm osm, templateUpdate *template.Template) error {
	descricao := strings.ReplaceAll(osm.ObservacaoReabertura, "<BR />", " ")
	descricao = strings.ReplaceAll(descricao, "  ", " ")
	if len(descricao) > 100 {
		descricao = descricao[0:100]
	}

	requestBody, err := executeTemplate(templateUpdate, map[string]interface{}{
		"WowoCode":           osm.Coswin,
		"WowoUserStatus":     "M",
		"WowoStatusComments": descricao,
	})
	if err != nil {
		return err
	}
	log.Printf("REABERTURA 1 - STATUS:\n----REQUEST---\n%v\n", requestBody.String())

	response, err := http.Post(
		CoswinURL+"/services/workorder/update",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Printf("----RESPONSE 1 - STATUS---\n%v\n", string(body))

	requestBody, err = executeTemplate(templateUpdate, map[string]interface{}{
		"WowoCode":        osm.Coswin,
		"WowoJobActivity": "<![CDATA[---OBSERVAÇÕES DE ANÁLISE---<br/>" + osm.ObservacaoAnalise + "<br/><br/>---OBSERVAÇÕES DE REABERTURA---<br/>" + osm.ObservacaoReabertura + "]]>",
		"WowoString12":    fmt.Sprintf("%v.%v", osm.Codigo, osm.Revisao),
	})
	if err != nil {
		return err
	}
	log.Printf("REABERTURA 2 - INFO:\n----REQUEST---\n%v\n", requestBody.String())

	response2, err := http.Post(
		CoswinURL+"/services/workorder/update",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response2.Body.Close()

	body, err = ioutil.ReadAll(response2.Body)
	if err != nil {
		return err
	}

	log.Printf("----RESPONSE 2 - INFO---\n%v\n", string(body))

	return nil
}
