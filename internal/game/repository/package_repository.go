package repository

import (
	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PackageRepository struct {
    db *gorm.DB
}


func NewPackageRepository(db *gorm.DB)*PackageRepository{
	return &PackageRepository{db: db}
}


// Cria objeto package no banco 
func (r *PackageRepository) Create( packages *entities.Package) (*entities.Package, error){
	packages.ID = uuid.NewString()
	if err := r.db.Create(packages).Error; err != nil {
        return nil, err
    }
	return  packages, nil
}


// cria um pacote com ID específico (para sincronização Raft)
func (r *PackageRepository) CreateWithID( packages *entities.Package) (*entities.Package, error) {
	result := r.db.Create(packages)
	if result.Error != nil {
		return nil, result.Error
	}
	return packages, nil
}	


// retorna todos os packages (para snapshot)
func (r *PackageRepository) GetAll() ([]*entities.Package, error) {
	var packages []*entities.Package
	result := r.db.Find(&packages)
	if result.Error != nil {
		return nil, result.Error
	}
	return packages, nil
}


// deleta todos os packages (para restore snapshot)
func (r *PackageRepository)DeleteAll() error {
	result := r.db.Exec("DELETE FROM packages")
	return  result.Error
}


func (r *PackageRepository) Delete(id string){
	r.db.Delete(id)
}



func (r *PackageRepository) FindById(id string)(*entities.Package, error){
	var packages entities.Package
	if err := r.db.Where("id = ?", id).First(&packages).Error; err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){ 
			return  nil, nil
		}
	return nil, err 
	
	}
	return  &packages, nil 
}



// Modifica status do pacote // "available", "locked", "opened"
func (r *PackageRepository) UpdatePackageStatus(packageID, status string) error {
    return r.db.Model(&entities.Package{}).
        Where("id = ?", packageID).
        Update("status", status).Error
}


func (r *PackageRepository) FindByIdWithCards(id string) (*entities.Package, error) {
    var pkg entities.Package
    err := r.db.Preload("Cards").First(&pkg, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &pkg, nil
}

