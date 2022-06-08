output "resource_group_name" {
  value = azurerm_resource_group.example.name
}

output "vnet_name" {
  value = module.network.vnet_name
}