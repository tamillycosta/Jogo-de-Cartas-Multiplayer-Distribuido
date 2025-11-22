const Card = artifacts.require("Card");
const PackageRegistry = artifacts.require("PackageRegistry");

module.exports = async function (deployer) {
    await deployer.deploy(Card);
    const card = await Card.deployed();

    // Agora sim: passa o endereço do contrato Card
    await deployer.deploy(PackageRegistry, card.address);
};
