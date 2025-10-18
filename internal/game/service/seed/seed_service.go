package seedService

import (

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	packageService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/packageService"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"log"
	"time"
)

// Responssavel por prencher o banco de dados com os pacotes / cartas
// apenas o servidor lider consegue fazer ter acessado a este serviço
type SeedService struct {
	raft           *raftService.RaftService
	packageService *packageService.PackageService
}

func New(raft *raftService.RaftService, pkgService *packageService.PackageService) *SeedService {
	return &SeedService{
		raft:           raft,
		packageService: pkgService,
	}
}

func  (s *SeedService) Init(allPackages []*entities.Package, err error){
	log.Println("Executando Seeds...")
	if err != nil {
		return
	}
	go func() {
		time.Sleep(3 * time.Second) 
		s.RunSeeds()

		s.startAutoReplenish(allPackages ,5 * time.Minute)
	}()
}

// executa seeds iniciais (apenas no líder)
func (s *SeedService) RunSeeds() {
	// Aguarda ser líder
	log.Println("[Seeds] Aguardando ser líder para executar seeds...")
	
	for i := 0; i < 30; i++ { 
		if s.raft.IsLeader() {
			log.Println("[Seeds] Sou líder! Executando seeds...")
			s.createInitialPackages()
			return
		}
		time.Sleep(1 * time.Second)
	}
	
	log.Println("[Seeds] Não sou líder, pulando seeds")
}

// cria pacotes iniciais
// quantidade inicial de 100 pacotes 
func (s *SeedService) createInitialPackages() {
	const initialPackageCount = 100 
	
	log.Printf("[Seeds] Criando %d pacotes iniciais...", initialPackageCount)
	
	successCount := 0
	failCount := 0
	
	for i := 0; i < initialPackageCount; i++ {
		err := s.packageService.CreatePackage()
		if err != nil {
			log.Printf("[Seeds] Erro ao criar pacote %d: %v", i+1, err)
			failCount++
		} else {
			successCount++
		}
		
		
		time.Sleep(50 * time.Millisecond)
	}
	
	log.Printf("[Seeds] Pacotes criados: %d sucesso, %d falhas", successCount, failCount)
}




// verifica se precisa criar mais pacotes
// em caso dos pacotes terem esgotado 
func (s *SeedService) CreatePackagesIfNeeded(allPackages []*entities.Package) {
	if !s.raft.IsLeader() {
		return
	}
	
	available, err := s.packageService.GetAvailablePackages()
	if err != nil {
		log.Printf("⚠️ [Seeds] Erro ao verificar pacotes: %v", err)
		return
	}
	
	const minPackages = 50
	
	if len(available) < minPackages {
		toCreate := minPackages - len(available)
		log.Printf("[Seeds] Criando %d novos pacotes (atual: %d)", toCreate, len(available))
		
		for i := 0; i < toCreate; i++ {
			s.packageService.CreatePackage()
			time.Sleep(50 * time.Millisecond)
		}
	}
}


// inicia reabastecimento automático de pacotes
func (s *SeedService) startAutoReplenish(allPackages []*entities.Package,interval time.Duration) {
	ticker := time.NewTicker(interval)
	
	go func() {
		for range ticker.C {
			s.CreatePackagesIfNeeded(allPackages)
		}
	}()
	
	log.Printf("[Seeds] Auto-replenish ativado (intervalo: %v)", interval)
}