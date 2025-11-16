// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract PackageRegistry {
    
    struct Package {
        bytes32 id;              // ID único do pacote
        bytes32[] cardIds;       // IDs das 5 cartas
        bool opened;             // Se foi aberto
        bytes32 openedBy;        // ID do jogador que abriu
        uint256 createdAt;       // Timestamp de criação
    }
    
    // Mapeamento de packageID -> Package
    mapping(bytes32 => Package) public packages;
    
    // Lista de todos os IDs (para iterar)
    bytes32[] public packageIds;
    
    // Eventos para o frontend ouvir
    event PackageCreated(bytes32 indexed packageId, bytes32[] cardIds, uint256 timestamp);
    event PackageOpened(bytes32 indexed packageId, bytes32 indexed playerId, uint256 timestamp);
    
    // ===== CRIAR PACOTE =====
    
    function createPackage(bytes32 _packageId, bytes32[] memory _cardIds) public {
        require(_cardIds.length == 5, "Pacote deve ter exatamente 5 cartas");
        require(packages[_packageId].createdAt == 0, "Pacote ja existe");
        
        packages[_packageId] = Package({
            id: _packageId,
            cardIds: _cardIds,
            opened: false,
            openedBy: bytes32(0),
            createdAt: block.timestamp
        });
        
        packageIds.push(_packageId);
        
        emit PackageCreated(_packageId, _cardIds, block.timestamp);
    }
    
    
    
    // ===== CONSULTAS  =====
    
    // Verifica se pacote existe
    function packageExists(bytes32 _packageId) public view returns (bool) {
        return packages[_packageId].createdAt != 0;
    }
    
    // Retorna informações do pacote
    function getPackage(bytes32 _packageId) public view returns (Package memory) {
        require(packages[_packageId].createdAt != 0, "Pacote nao existe");
        return packages[_packageId];
    }
    
    // Total de pacotes criados
    function getTotalPackages() public view returns (uint256) {
        return packageIds.length;
    }
    
    // Pegar pacote por índice (para iteração)
    function getPackageByIndex(uint256 index) public view returns (Package memory) {
        require(index < packageIds.length, "Index invalido");
        bytes32 packageId = packageIds[index];
        return packages[packageId];
    }
    
    // Listar últimos N pacotes
    function getRecentPackages(uint256 count) public view returns (Package[] memory) {
        uint256 total = packageIds.length;
        uint256 resultCount = count > total ? total : count;
        
        Package[] memory result = new Package[](resultCount);
        
        for (uint256 i = 0; i < resultCount; i++) {
            bytes32 packageId = packageIds[total - 1 - i];
            result[i] = packages[packageId];
        }
        
        return result;
    }
}