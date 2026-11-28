resource "azurerm_resource_group" "main" {
  name     = "demo-azure-https"
  location = var.location
}

resource "azurerm_virtual_network" "main" {
  name                = "demo-network"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  address_space       = ["172.16.0.0/24"]
}

resource "azurerm_subnet" "main" {
  name                 = "demo-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["172.16.0.0/24"]
}

# Traffic inside the VNet is allowed by the default AllowVnetInBound rule, and
# inbound from the internet is denied by the default DenyAllInBound rule.
resource "azurerm_network_security_group" "main" {
  name                = "demo-nsg"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  # Allow SSH from Azure Bastion (Developer SKU) so the VMs (no public IP) can be
  # reached for debugging. The Developer SKU connects from 168.63.129.16.
  security_rule {
    name                       = "AllowBastionSsh"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "168.63.129.16"
    destination_address_prefix = "*"
  }
}

resource "azurerm_subnet_network_security_group_association" "main" {
  subnet_id                 = azurerm_subnet.main.id
  network_security_group_id = azurerm_network_security_group.main.id
}

# New Azure subnets have no default outbound access. The NAT gateway lets VMs
# without public IPs reach the internet.
resource "azurerm_public_ip" "nat" {
  name                = "demo-nat-ip"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Standard"
  allocation_method   = "Static"
}

resource "azurerm_nat_gateway" "main" {
  name                = "demo-nat"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_name            = "Standard"
}

resource "azurerm_nat_gateway_public_ip_association" "main" {
  nat_gateway_id       = azurerm_nat_gateway.main.id
  public_ip_address_id = azurerm_public_ip.nat.id
}

resource "azurerm_subnet_nat_gateway_association" "main" {
  subnet_id      = azurerm_subnet.main.id
  nat_gateway_id = azurerm_nat_gateway.main.id
}

# Break-glass SSH access from the Azure portal (Connect > Bastion). The Developer
# SKU is free and needs no dedicated subnet or public IP.
resource "azurerm_bastion_host" "main" {
  name                = "demo-bastion"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Developer"
  virtual_network_id  = azurerm_virtual_network.main.id
}

# Azure requires an admin user on every Linux VM. This key is used for Bastion
# logins: terraform output -raw vm_admin_private_key
resource "tls_private_key" "vm_admin" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

locals {
  vm_admin_username = "azureadmin"

  debian_image = {
    publisher = "Debian"
    offer     = "debian-13"
    sku       = "13-gen2"
    version   = "latest"
  }
}
