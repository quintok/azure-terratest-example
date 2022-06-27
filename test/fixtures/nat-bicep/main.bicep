@secure()
param adminPassword string

param location string

module network '../../../network.bicep' = {
  name: 'network'
  params: {
    location: location
  }
}

resource nic 'Microsoft.Network/networkInterfaces@2021-08-01' = {
  name: 'test-machine-nic'
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'testconfiguration1'
        properties: {
          privateIPAllocationMethod: 'Dynamic'
          subnet: {
            id: network.outputs.privateSubnetId
          }
        }
      }
    ]
  }
}

resource vm 'Microsoft.Compute/virtualMachines@2022-03-01' = {
  name: 'test-machine'
  location: location
  properties: {
     hardwareProfile: {
      vmSize: 'Standard_B2ms'
     }
     networkProfile: {
      networkInterfaces: [
        {
          id: nic.id
          properties: {
            deleteOption: 'Delete'
            primary: true
          }
        }
      ]
     }
     osProfile: {
      adminPassword: adminPassword
      adminUsername: 'azureuser'
      computerName: 'test-machine'
     }
     storageProfile: {
      imageReference: {
        publisher: 'Canonical'
        offer: 'UbuntuServer'
        sku: '16.04-LTS'
        version: 'latest'
      }
      osDisk: {
        createOption: 'FromImage'
      }
     }
  }
}

resource vmExtension 'Microsoft.Compute/virtualMachines/extensions@2022-03-01' = {
  name: '${vm.name}/AzurePolicyforLinux'
  location: location
  properties: {
    publisher: 'Microsoft.GuestConfiguration'
    type: 'ConfigurationForLinux'
    typeHandlerVersion: '1.0'
    autoUpgradeMinorVersion: true
    enableAutomaticUpgrade: true
    
  }
}
