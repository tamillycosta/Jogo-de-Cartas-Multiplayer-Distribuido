// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./Card.sol";

contract PackageRegistry {
    
    Card public card;
    
    // ===== ESTRUTURAS =====
    
    struct Package {
        string id;              
        string[] cardIds;      
        bool opened;
        string openedBy;      
        uint256 createdAt;
    }
    
    // ===== STORAGE =====
    
  
    mapping(string => Package) public packages;
    string[] public packageIds;
    
    // ===== EVENTOS =====
    
    event PackageCreated(
        string packageId,      
        string[] cardIds,      
        uint256 timestamp
    );
    
    event PackageOpened(
        string packageId,      
        string playerId,        
        address indexed playerAddress,
        uint256 timestamp
    );
    
    // ===== CONSTRUCTOR =====
    
    constructor(address _cardAddress) {
        require(_cardAddress != address(0), "Endereco da carta invalido");
        card = Card(_cardAddress);
    }
    
    // ===== CRIAR PACOTE =====
    
    function createPackage(
        string memory _packageId,  
        string[] memory _cardIds   
    ) public {
        require(_cardIds.length == 5, "Pacote deve ter 5 cartas");
        require(packages[_packageId].createdAt == 0, "Pacote ja existe");
        require(bytes(_packageId).length > 0, "PackageId vazio");
        
        packages[_packageId] = Package({
            id: _packageId,
            cardIds: _cardIds,
            opened: false,
            openedBy: "",
            createdAt: block.timestamp
        });
        
        packageIds.push(_packageId);
        
        emit PackageCreated(_packageId, _cardIds, block.timestamp);
    }
    
    // ===== ABRIR PACOTE =====
    
    function openPackage(
        string memory _packageId,    
        string memory _playerId,    
        address _playerAddress,
        string[] memory _templateIds  
    ) public {
        Package storage pkg = packages[_packageId];
        
        require(pkg.createdAt != 0, "Pacote nao existe");
        require(!pkg.opened, "Pacote ja foi aberto");
        require(_templateIds.length == pkg.cardIds.length, "Templates incorretos");
        require(_playerAddress != address(0), "Endereco invalido");
        
        pkg.opened = true;
        pkg.openedBy = _playerId;
        
        for (uint256 i = 0; i < pkg.cardIds.length; i++) {
            card.mintCard(
                pkg.cardIds[i],
                _templateIds[i],
                _packageId,
                _playerAddress
            );
        }
        
        emit PackageOpened(_packageId, _playerId, _playerAddress, block.timestamp);
    }
    
    // ===== CONSULTAS =====
    
    function packageExists(string memory _packageId) public view returns (bool) {
        return packages[_packageId].createdAt != 0;
    }
    
    function getPackage(string memory _packageId) public view returns (Package memory) {
        require(packages[_packageId].createdAt != 0, "Pacote nao existe");
        return packages[_packageId];
    }
    
    function getTotalPackages() public view returns (uint256) {
        return packageIds.length;
    }
    
    function getPackageByIndex(uint256 index) public view returns (Package memory) {
        require(index < packageIds.length, "Index invalido");
        string memory packageId = packageIds[index];
        return packages[packageId];
    }
    
    function getRecentPackages(uint256 count) public view returns (Package[] memory) {
        uint256 total = packageIds.length;
        uint256 resultCount = count > total ? total : count;
        
        Package[] memory result = new Package[](resultCount);
        
        for (uint256 i = 0; i < resultCount; i++) {
            string memory packageId = packageIds[total - 1 - i];
            result[i] = packages[packageId];
        }
        
        return result;
    }
}