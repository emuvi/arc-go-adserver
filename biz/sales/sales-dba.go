package sales

import (
	"adserver/biz/company"

	"adserver/motor"
)

func GetLastPreOrders(onTransit *motor.Convey) bool {
	if !company.GetPersonOfUser(onTransit) {
		goto BadError
	}
	if !onTransit.Query(
		`SELECT 
			prepedidos.enviado_data AS sent,
			cliente.nome AS name, 
			cliente.fantasia AS fantasy,
			prepedidos.total AS total
		FROM 
			prepedidos 
		JOIN 
			pessoas AS cliente
			ON cliente.codigo = prepedidos.cliente
		WHERE 
			prepedidos.representante LIKE $1
		ORDER BY 
			prepedidos.enviado_data DESC,
			prepedidos.enviado_hora DESC
		LIMIT 7`, onTransit.Get("PersonOfUser"),
	) {
		goto BadError
	}
	if !onTransit.PutRows("LasPreOrders",
		motor.Fetcher{Column: "sent", Form: &motor.Formatter{Type: motor.FormatDate}},
		motor.Fetcher{Column: "name"},
		motor.Fetcher{Column: "fantasy"},
		motor.Fetcher{Column: "total", Form: &motor.Formatter{Type: motor.FormatCurrency}},
	) {
		goto BadError
	}
	return true
BadError:
	onTransit.PutError("can't get the last pre-orders")
	return false
}
