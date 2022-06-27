param location string

resource vnet 'Microsoft.Network/virtualNetworks@2021-05-01' = {
  name: uniqueString(resourceGroup().id)
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.0.0.0/16'
      ]
    }
    subnets: [
      {
        name: 'private'
        properties: {
          natGateway: {
            id: nat.id
          }
          addressPrefix: '10.0.0.0/24'
          privateEndpointNetworkPolicies: 'Enabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
        }
      }
    ]
  }
}

resource natPublicPrefix 'Microsoft.Network/publicIPPrefixes@2021-08-01' = {
  name: 'nat-gateway-publicIPPrefix'
  location: location
  sku: {
    name: 'Standard'
  }
  properties: {
    prefixLength: 30
  }
}

resource natPublicIP 'Microsoft.Network/publicIPAddresses@2021-08-01' = {
  name: 'nat-gateway-publicIP'
  location: location
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
  }
}

resource nat 'Microsoft.Network/natGateways@2021-08-01' = {
  location: location
  name: 'example-nat'
  sku: {
    name: 'Standard'
  }
  properties: {
    idleTimeoutInMinutes: 10
    publicIpAddresses: [
      {
        id: natPublicIP.id
      }
    ]
    publicIpPrefixes: [
      {
        id: natPublicPrefix.id
      }
    ]
  }
}

output privateSubnetId string = vnet.properties.subnets[0].id
