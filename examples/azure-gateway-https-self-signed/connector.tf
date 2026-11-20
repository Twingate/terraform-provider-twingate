resource "azurerm_network_interface" "connector" {
  name                = "demo-connector"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_linux_virtual_machine" "connector" {
  name                  = "demo-connector"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  size                  = var.vm_size
  network_interface_ids = [azurerm_network_interface.connector.id]

  admin_username = local.vm_admin_username
  admin_ssh_key {
    username   = local.vm_admin_username
    public_key = tls_private_key.vm_admin.public_key_openssh
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = local.debian_image.publisher
    offer     = local.debian_image.offer
    sku       = local.debian_image.sku
    version   = local.debian_image.version
  }

  custom_data = base64encode(templatefile("${path.module}/scripts/connector-startup.sh", {
    access_token  = twingate_connector_tokens.main.access_token
    refresh_token = twingate_connector_tokens.main.refresh_token
    twingate_url  = "https://${var.tg_network}.${var.tg_url}"
  }))

  # The startup script runs once at boot and downloads the Connector, so the
  # outbound path (subnet -> NAT gateway -> internet) must exist before the VM
  # boots.
  depends_on = [
    azurerm_subnet_nat_gateway_association.main,
    azurerm_nat_gateway_public_ip_association.main,
  ]
}
