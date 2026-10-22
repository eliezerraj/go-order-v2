//
//WHO    = Principal (client_id)
//WHAT   = Action (inventory create)
//WHICH  = Resource (order order) when empty means all
//

permit(
    principal,
    action == Action::"get:time-series-order-items",
    resource
)
when {
    principal.scopes.contains(
        "order:history"
    )
};

permit(
    principal,
    action == Action::"get:order",
    resource
)
when {
    principal.scopes.contains(
        "order:read"
    )
};

permit(
    principal,
    action == Action::"add:order",
    resource
)
when {
    principal.scopes.contains(
        "order:create"
    )
};

permit(
    principal,
    action == Action::"get:checkou",
    resource
)
when {
    principal.scopes.contains(
        "checkout:read"
    )
};

permit(
    principal,
    action == Action::"add:checkout",
    resource
)
when {
    principal.scopes.contains(
        "checkout:create"
    )
};