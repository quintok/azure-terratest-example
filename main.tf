provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "group" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_virtual_network" "network" {
  name                = "example-network"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "public" {
  name                 = "public-subnet"
  resource_group_name  = azurerm_resource_group.group.name
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_subnet" "private" {
  name                 = "private-subnet"
  resource_group_name  = azurerm_resource_group.group.name
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "nat-ip" {
  name                = "nat-gateway-publicIP"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_public_ip_prefix" "nat-ip-prefix" {
  name                = "nat-gateway-publicIPPrefix"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  prefix_length       = 30
}

resource "azurerm_nat_gateway" "nat" {
  name                    = "example-nat"
  location                = azurerm_resource_group.group.location
  resource_group_name     = azurerm_resource_group.group.name
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
}

resource "azurerm_nat_gateway_public_ip_association" "example" {
  nat_gateway_id       = azurerm_nat_gateway.nat.id
  public_ip_address_id = azurerm_public_ip.nat-ip.id
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "example" {
  nat_gateway_id      = azurerm_nat_gateway.nat.id
  public_ip_prefix_id = azurerm_public_ip_prefix.nat-ip-prefix.id
}

resource "azurerm_subnet_nat_gateway_association" "association" {
  subnet_id      = azurerm_subnet.public.id
  nat_gateway_id = azurerm_nat_gateway.nat.id
}

resource "azurerm_subnet_nat_gateway_association" "example" {
  subnet_id      = azurerm_subnet.private.id
  nat_gateway_id = azurerm_nat_gateway.nat.id
}

resource "azurerm_network_security_group" "private-subnet" {
  name                = "private-subnet"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
}

resource "azurerm_network_security_rule" "block-internet-inbound" {
  name                        = "block-internet"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Deny"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "Internet"
  destination_address_prefix  = "VirtualNetwork"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.private-subnet.name
}

resource "azurerm_network_security_rule" "block-internet-outbound" {
  name                        = "block-internet"
  priority                    = 101
  direction                   = "Outbound"
  access                      = "Deny"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "VirtualNetwork"
  destination_address_prefix  = "Internet"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.private-subnet.name
}

resource "azurerm_subnet_network_security_group_association" "private-subnet" {
  subnet_id                 = azurerm_subnet.private.id
  network_security_group_id = azurerm_network_security_group.private-subnet.id

  depends_on = [
    azurerm_network_security_rule.block-internet-outbound,
    azurerm_network_security_rule.block-internet-inbound
  ]
}