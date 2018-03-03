export class Player
{
    Name: string;
    Color: string;

    constructor(playerObject: Object)
    {
        this.Name = playerObject["Name"];
        this.Color = playerObject["Color"];
    }
}