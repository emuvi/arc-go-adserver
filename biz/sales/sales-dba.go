package sales

import (
	"adserver/biz/company"

	"adserver/motor"
)

func GetLastPreOrders(onTransit *motor.Convey) bool {
	var result []map[string]interface{}
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
		LIMIT 7`, onTransit.Get("PersonOfUser")) {
		goto BadError
	}
	for onTransit.Next() {
		values, _ := rows.Values()
		if len(values) < 4 {
			onTransit.PutError("can't get the last pre-orders. (wrong number of columns)")
			return false
		}
		row := map[string]interface{}{}
		row["sent"] = onTransit.FormatDate(values[0])
		row["name"] = values[1]
		row["fantasy"] = values[2]
		row["total"] = onTransit.FormatCurrency(values[3])
		result = append(result, row)
	}
	onTransit.Set("LastPreOrders", result)
	return true
BadError:
	onTransit.PutError("can't get the last pre-orders")
	return false
}
