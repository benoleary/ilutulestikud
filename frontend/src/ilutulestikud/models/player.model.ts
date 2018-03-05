export class Player
{
    Identifier: string;
    Name: string;
    Color: string;

    constructor(playerObject: Object)
    {
        this.Identifier = playerObject["Identifier"];
        this.Name = playerObject["Name"];
        this.Color = playerObject["Color"];
    }
}