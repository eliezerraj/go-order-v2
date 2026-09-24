# go-order-v2
go-order-v2


```sql
select 	oi.fk_product_id, 
		created_at::date AS date_part,
		EXTRACT(HOUR FROM created_at) AS hour_part,
		count(oi.fk_product_id) as count,
		sum (oi.quantity) as sum_quantity,
		sum(oi.amount) as sum_amount,
		sum(oi.discount) as sum_discount
from order_item oi
where oi.fk_product_id = 26
group by oi.fk_product_id, date_part ,hour_part
order by date_part ,hour_part asc
limit 7 offset 0
´´´