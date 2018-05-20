export class BackendIdentification
{
    NameForPost: string;
    IdentifierForGet: string;

    constructor(NameForPost: string, IdentifierForGet: string)
    {
        this.NameForPost = NameForPost;
        this.IdentifierForGet = IdentifierForGet;
    }

    static FromObject(objectWithNameAndIdentifier: Object): BackendIdentification
    {
        return new BackendIdentification(
            objectWithNameAndIdentifier["Name"],
            objectWithNameAndIdentifier["Identifier"]);
    }
}