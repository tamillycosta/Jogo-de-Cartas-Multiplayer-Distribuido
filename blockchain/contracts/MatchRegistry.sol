// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";

contract MatchRegistry is Ownable {
    
    // ===== ENUMS =====
    
    enum MatchType { LOCAL, REMOTE }
    enum MatchStatus { IN_PROGRESS, FINISHED, ABANDONED }
    
    // ===== ESTRUTURAS =====
    
    struct Match {
        string matchId;           
        MatchType matchType;     
        MatchStatus status;      
        string player1Id;        
        string player2Id;         
        string winnerId;          
        uint256 startedAt;        
        uint256 finishedAt;     
        uint256 totalTurns;      
        string serverHost;        
    }
    
    // ===== STORAGE =====
    
    mapping(string => Match) public matches;
    string[] public matchIds;
    
    // Estatísticas por jogador
    mapping(string => uint256) public playerTotalMatches;
    mapping(string => uint256) public playerWins;
    mapping(string => uint256) public playerLosses;
    
    // Estatísticas gerais
    uint256 public totalMatches;
    uint256 public totalLocalMatches;
    uint256 public totalRemoteMatches;
    uint256 public activeMatches;
    uint256 public finishedMatches;
    
    // ===== EVENTOS =====
    
    event MatchStarted(
        string indexed matchId,
        MatchType matchType,
        string player1Id,
        string player2Id,
        uint256 timestamp
    );
    
    event MatchFinished(
        string indexed matchId,
        string winnerId,
        uint256 totalTurns,
        uint256 timestamp
    );
    
    event MatchAbandoned(
        string indexed matchId,
        string playerId,
        uint256 timestamp
    );
    
    // ===== CONSTRUCTOR =====
    
    constructor() Ownable(msg.sender) {}
    
    // ===== REGISTRAR PARTIDA =====
    
    function startMatch(
        string memory _matchId,
        MatchType _matchType,
        string memory _player1Id,
        string memory _player2Id,
        string memory _serverHost
    ) public {
        require(matches[_matchId].startedAt == 0, "Partida ja existe");
        require(bytes(_player1Id).length > 0, "Player1 invalido");
        require(bytes(_player2Id).length > 0, "Player2 invalido");
        
        matches[_matchId] = Match({
            matchId: _matchId,
            matchType: _matchType,
            status: MatchStatus.IN_PROGRESS,
            player1Id: _player1Id,
            player2Id: _player2Id,
            winnerId: "",
            startedAt: block.timestamp,
            finishedAt: 0,
            totalTurns: 0,
            serverHost: _serverHost
        });
        
        matchIds.push(_matchId);
        
        totalMatches++;
        activeMatches++;
        
        if (_matchType == MatchType.LOCAL) {
            totalLocalMatches++;
        } else {
            totalRemoteMatches++;
        }
        
        playerTotalMatches[_player1Id]++;
        playerTotalMatches[_player2Id]++;
        
        emit MatchStarted(_matchId, _matchType, _player1Id, _player2Id, block.timestamp);
    }
    
    // ===== FINALIZAR PARTIDA =====
    
    function finishMatch(
        string memory _matchId,
        string memory _winnerId,
        uint256 _totalTurns
    ) public {
        Match storage matchData = matches[_matchId];
        
        require(matchData.startedAt != 0, "Partida nao existe");
        require(matchData.status == MatchStatus.IN_PROGRESS, "Partida nao esta em andamento");
        require(
            keccak256(bytes(_winnerId)) == keccak256(bytes(matchData.player1Id)) ||
            keccak256(bytes(_winnerId)) == keccak256(bytes(matchData.player2Id)),
            "Vencedor invalido"
        );
        
        matchData.status = MatchStatus.FINISHED;
        matchData.winnerId = _winnerId;
        matchData.finishedAt = block.timestamp;
        matchData.totalTurns = _totalTurns;
        
        activeMatches--;
        finishedMatches++;
        
        // Atualiza estatísticas dos jogadores
        playerWins[_winnerId]++;
        
        string memory loserId = keccak256(bytes(_winnerId)) == keccak256(bytes(matchData.player1Id))
            ? matchData.player2Id
            : matchData.player1Id;
        
        playerLosses[loserId]++;
        
        emit MatchFinished(_matchId, _winnerId, _totalTurns, block.timestamp);
    }
    
    // ===== PARTIDA ABANDONADA =====
    
    function abandonMatch(
        string memory _matchId,
        string memory _playerId
    ) public {
        Match storage matchData = matches[_matchId];
        
        require(matchData.startedAt != 0, "Partida nao existe");
        require(matchData.status == MatchStatus.IN_PROGRESS, "Partida nao esta em andamento");
        
        matchData.status = MatchStatus.ABANDONED;
        matchData.finishedAt = block.timestamp;
        
        // Quem não abandonou vence
        if (keccak256(bytes(_playerId)) == keccak256(bytes(matchData.player1Id))) {
            matchData.winnerId = matchData.player2Id;
            playerWins[matchData.player2Id]++;
            playerLosses[_playerId]++;
        } else {
            matchData.winnerId = matchData.player1Id;
            playerWins[matchData.player1Id]++;
            playerLosses[_playerId]++;
        }
        
        activeMatches--;
        finishedMatches++;
        
        emit MatchAbandoned(_matchId, _playerId, block.timestamp);
    }
    
    // ===== CONSULTAS =====
    
    function getMatch(string memory _matchId) public view returns (Match memory) {
        require(matches[_matchId].startedAt != 0, "Partida nao existe");
        return matches[_matchId];
    }
    
    function matchExists(string memory _matchId) public view returns (bool) {
        return matches[_matchId].startedAt != 0;
    }
    
    function getTotalMatches() public view returns (uint256) {
        return totalMatches;
    }
    
    function getMatchByIndex(uint256 index) public view returns (Match memory) {
        require(index < matchIds.length, "Index invalido");
        string memory matchId = matchIds[index];
        return matches[matchId];
    }
    
    function getPlayerStats(string memory _playerId) public view returns (
        uint256 total,
        uint256 wins,
        uint256 losses,
        uint256 winRate
    ) {
        total = playerTotalMatches[_playerId];
        wins = playerWins[_playerId];
        losses = playerLosses[_playerId];
        
        if (total > 0) {
            winRate = (wins * 100) / total;
        } else {
            winRate = 0;
        }
        
        return (total, wins, losses, winRate);
    }
    
    function getRecentMatches(uint256 count) public view returns (Match[] memory) {
        uint256 total = matchIds.length;
        uint256 resultCount = count > total ? total : count;
        
        Match[] memory result = new Match[](resultCount);
        
        for (uint256 i = 0; i < resultCount; i++) {
            string memory matchId = matchIds[total - 1 - i];
            result[i] = matches[matchId];
        }
        
        return result;
    }
    
    function getSystemStats() public view returns (
        uint256 _totalMatches,
        uint256 _totalLocal,
        uint256 _totalRemote,
        uint256 _active,
        uint256 _finished
    ) {
        return (
            totalMatches,
            totalLocalMatches,
            totalRemoteMatches,
            activeMatches,
            finishedMatches
        );
    }
}