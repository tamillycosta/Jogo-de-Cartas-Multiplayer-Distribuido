// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract Card is ERC721, Ownable {
    
    // ===== ESTRUTURAS =====
    
    struct CardMetadata {
        string cardId;         
        string  templateId;      
        string  packageId;      
        uint256 mintedAt;     
        
    }
    
    // ===== STORAGE =====
    
    // tokenId => metadados da carta 
    mapping(uint256 => CardMetadata) public cards;
    
    // cardId  => tokenId (NFT)
    mapping(string => uint256) public cardIdToTokenId;
    
    // Rastreamento das cartas de cada jogador
    mapping(address => uint256[]) private _playerTokens;
    mapping(uint256 => uint256) private _playerTokensIndex;
    
 
    uint256 private _tokenIdCounter;
    
    // ===== EVENTOS =====
    
    event CardMinted(
        uint256 indexed tokenId, 
        string indexed cardId, 
        address indexed owner
    );
    
    event CardTransferred(
        uint256 indexed tokenId, 
        string indexed cardId, 
        address from, 
        address to
    );
    
    // ===== CONSTRUCTOR =====
    
    constructor() ERC721("GameCard", "CARD") Ownable(msg.sender) {}
    
    // ===== MINT (CRIAR CARTA COMO NFT) =====
    
    function mintCard(
        string  memory _cardId,
        string memory  _templateId,
        string memory _packageId,
        address _initialOwner
    ) public returns (uint256) {
        require(cardIdToTokenId[_cardId] == 0, "Carta ja existe");
        require(_initialOwner != address(0), "Endereco invalido");
        
        _tokenIdCounter++;
        uint256 newTokenId = _tokenIdCounter;
        
        // Cria o NFT - ERC-721 rastreia o dono automaticamente
        _safeMint(_initialOwner, newTokenId);
        
        // Salva metadados  
        cards[newTokenId] = CardMetadata({
            cardId: _cardId,
            templateId: _templateId,
            packageId: _packageId,
            mintedAt: block.timestamp
        });
        
        // Índice para buscar tokenId pelo cardId
        cardIdToTokenId[_cardId] = newTokenId;
        
        emit CardMinted(newTokenId, _cardId, _initialOwner);
        
        return newTokenId;
    }
    
    // ===== TRANSFERIR CARTA =====
    
    function transferCard(
        uint256 _tokenId,
        address _to
    ) public {
        require(_to != address(0), "Endereco invalido");
        
        address currentOwner = ownerOf(_tokenId);
        
        // Verifica permissões
        require(
            currentOwner == msg.sender || 
            getApproved(_tokenId) == msg.sender ||
            isApprovedForAll(currentOwner, msg.sender),
            "Sem permissao para transferir"
        );
        
        // ERC-721 atualiza o dono automaticamente
        _transfer(currentOwner, _to, _tokenId);
    }
    
    // ===== CONSULTAS =====
    
    // Retorna o dono ATUAL de uma carta pelo cardId
    function getCardOwner(string memory _cardId) public view returns (address) {
        uint256 tokenId = cardIdToTokenId[_cardId];
        require(tokenId != 0, "Carta nao existe");
        
     
        return ownerOf(tokenId);
    }
    
    // Retorna todos os tokenIds das cartas de um jogador
    function getPlayerCards(address _player) public view returns (uint256[] memory) {
        require(_player != address(0), "Endereco invalido");
        return _playerTokens[_player];
    }
    
    // Retorna metadados de uma carta (dados imutáveis)
    function getCardMetadata(uint256 _tokenId) public view returns (
        string  memory _cardId,
        string memory  _templateId,
        string memory _packageId,
        uint256 mintedAt,
        address currentOwner
    ) {
        require(_exists(_tokenId), "Carta nao existe");
        
        CardMetadata memory metadata = cards[_tokenId];
        
        return (
            metadata.cardId,
            metadata.templateId,
            metadata.packageId,
            metadata.mintedAt,
            ownerOf(_tokenId)  
        );
    }
    
    // Total de cartas criadas
    function getTotalCards() public view returns (uint256) {
        return _tokenIdCounter;
    }
    
   
    function _exists(uint256 tokenId) internal view returns (bool) {
        try this.ownerOf(tokenId) returns (address owner) {
            return owner != address(0);
        } catch {
            return false;
        }
    }
    
    // ===== OVERRIDE DO ERC-721 =====
    
    // Sobrescreve para manter o índice de cartas por jogador atualizado
    function _update(
        address to,
        uint256 tokenId,
        address auth
    ) internal virtual override returns (address) {
        address from = super._update(to, tokenId, auth);
        
        // Remove do array do dono anterior
        if (from != address(0) && from != to) {
            _removeTokenFromOwnerEnumeration(from, tokenId);
        }
        
        // Adiciona ao array do novo dono
        if (to != address(0) && from != to) {
            _addTokenToOwnerEnumeration(to, tokenId);
        }
        
        // Emite evento customizado para transferências
        if (from != address(0) && to != address(0)) {
            emit CardTransferred(tokenId, cards[tokenId].cardId, from, to);
        }
        
        return from;
    }
    
    // ===== HELPERS =====
    
    function _addTokenToOwnerEnumeration(address to, uint256 tokenId) private {
        _playerTokensIndex[tokenId] = _playerTokens[to].length;
        _playerTokens[to].push(tokenId);
    }
    
    function _removeTokenFromOwnerEnumeration(address from, uint256 tokenId) private {
        uint256 lastTokenIndex = _playerTokens[from].length - 1;
        uint256 tokenIndex = _playerTokensIndex[tokenId];

        if (tokenIndex != lastTokenIndex) {
            uint256 lastTokenId = _playerTokens[from][lastTokenIndex];
            _playerTokens[from][tokenIndex] = lastTokenId;
            _playerTokensIndex[lastTokenId] = tokenIndex;
        }

        _playerTokens[from].pop();
        delete _playerTokensIndex[tokenId];
    }
}