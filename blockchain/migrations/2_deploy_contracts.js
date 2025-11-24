const Card = artifacts.require("Card");
const PackageRegistry = artifacts.require("PackageRegistry");
const MatchRegistry = artifacts.require("MatchRegistry");


module.exports = async function (deployer) {

    await deployer.deploy(Card);
    const card = await Card.deployed();

    await deployer.deploy(PackageRegistry, card.address);
    const pack = await PackageRegistry.deployed()

    await deployer.deploy(MatchRegistry);
    const match = await MatchRegistry.deployed();

    console.log(" Card deployed:", card.address);
    console.log(" PackageRegistry deployed:", pack.address);
    console.log(" MatchRegistry deployed:", match.address);
};
