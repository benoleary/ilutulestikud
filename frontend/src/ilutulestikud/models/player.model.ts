import { BackendIdentification } from './backendidentification.model';

export class Player
{
    ForBackend: BackendIdentification;
    Color: string;

    constructor(playerObject: Object)
    {
        this.ForBackend = BackendIdentification.FromObject(playerObject);
        this.Color = playerObject["Color"];
    }

    Name(): string
    {
        return this.ForBackend.NameForPost
    }
}