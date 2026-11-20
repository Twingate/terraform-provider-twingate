locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })
}

resource "azurerm_network_interface" "gateway" {
  name                = "demo-gateway"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_linux_virtual_machine" "gateway" {
  name                  = "demo-gateway"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  size                  = var.vm_size
  network_interface_ids = [azurerm_network_interface.gateway.id]

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

  # cloud-init only runs custom_data on first boot. Changing custom_data forces a
  # new VM, so a new certificate, key or config replaces the VM instead of
  # updating it in place.
  custom_data = base64encode(templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = local.gateway_config
  }))

  # The startup script runs once at boot and downloads the Gateway binary, so the
  # outbound path (subnet -> NAT gateway -> internet) must exist before the VM
  # boots.
  depends_on = [
    azurerm_subnet_nat_gateway_association.main,
    azurerm_nat_gateway_public_ip_association.main,
  ]
}
