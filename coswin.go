package main

import (
	"encoding/xml"
	"io/ioutil"
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
		"WowoString12":    osm.Codigo,
		"WowoNumber12":    osm.Revisao,
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

	envFind := envelopeFind{}
	if err := xml.Unmarshal(body, &envFind); err != nil {
		return err
	}
	for _, wowo := range envFind.Body.RetworkorderfindfindParameters.WorkorderfindList.Workorderfind {
		if wowo.WowoString12 == "" || wowo.WowoUserStatus != "N" {
			continue
		}

		var osmAlvo *osm
		for _, osm := range listaOSM {
			if strconv.Itoa(osm.Codigo) == wowo.WowoString12 {
				osmAlvo = &osm
				break
			}
		}
		if osmAlvo == nil {
			continue
		}

		if osmAlvo.Revisao > wowo.WowoNumber12 {
			return reaberturaCoswin(*osmAlvo, templateUpdate)
		}

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
		"WowoJobActivity":    "<![CDATA[---OBSERVAÇÕES DE ANÁLISE---<br/>" + osm.ObservacaoAnalise + "<br/><br/>---OBSERVAÇÕES DE REABERTURA---<br/>" + osm.ObservacaoReabertura + "]]>",
		"WowoStatusComments": descricao,
		"WowoNumber12":       osm.Revisao,
	})
	if err != nil {
		return err
	}

	response, err := http.Post(
		CoswinURL+"/services/workorder/update",
		"application/xml",
		&requestBody,
	)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}
