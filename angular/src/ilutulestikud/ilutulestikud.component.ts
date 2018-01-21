import { Component } from '@angular/core';
import { OnInit } from '@angular/core/src/metadata/lifecycle_hooks';
import { IlutulestikudService } from './ilutulestikud.service';
import { Player } from './models/player.model'

@Component({
  selector: 'app-ilutulestikud',
  templateUrl: './ilutulestikud.component.html'
})
export class IlutulestikudComponent implements OnInit
{
  ilutulestikudService: IlutulestikudService;
  selectedPlayerName: string;
  registeredPlayerNames: string[];
  selectedPlayer: Player;
  registeredPlayers: Player[];
  isAddPlayerDialogueVisible: boolean;
  newPlayerName: string;
  informationText: string;
  

  constructor(ilutulestikudService: IlutulestikudService)
  {
    this.ilutulestikudService = ilutulestikudService;
    this.selectedPlayerName = null;
    this.registeredPlayerNames = [];
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.isAddPlayerDialogueVisible = false;
    this.newPlayerName = null;
    this.informationText = null;
  }

  ngOnInit(): void
  {
    this.ilutulestikudService.registeredPlayerNames().subscribe(
      fetchedPlayerNamesObject => this.parsePlayerNames(fetchedPlayerNamesObject),
      thrownError => this.handleError(thrownError),
      () => {});
    this.ilutulestikudService.registeredPlayers().subscribe(
      fetchedPlayersObject => this.parsePlayers(fetchedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }

  selectedPlayerText(): string
  {
    if (!this.selectedPlayerName)
    {
      return "No player selected yet";
    }

    return "Player: " + this.selectedPlayerName;
  }

  parsePlayers(fetchedPlayersObject: Object): void
  {
    console.log("fetchedPlayersObject = " + JSON.stringify(fetchedPlayersObject));
    this.registeredPlayers.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const playerObject of fetchedPlayersObject["Players"]) {
      console.log("playerObject = " + JSON.stringify(playerObject));
      this.registeredPlayers.push(new Player(playerObject["Name"], playerObject["Color"]));
    }

    console.log("this.registeredPlayers = " + JSON.stringify(this.registeredPlayers));
  }

  parsePlayerNames(fetchedPlayerNamesObject: Object): void
  {
    this.registeredPlayerNames = fetchedPlayerNamesObject["Names"].slice();
  }

  showAddPlayerDialogue(): void
  {
    this.newPlayerName = "";
    this.isAddPlayerDialogueVisible = true;
  }

  cancelAddPlayerInDialogue(): void {
    this.informationText = null;
    this.isAddPlayerDialogueVisible = false;
  }

  addPlayerFromDialogue(): void
  {
    this.informationText = null;
    this.ilutulestikudService.newPlayer(this.newPlayerName).subscribe(
      returnedPlayerNamesObject => this.parsePlayerNames(returnedPlayerNamesObject),
      thrownError => this.handleError(thrownError),
      () => {});
    this.isAddPlayerDialogueVisible = false;
  }

  handleError(thrownError: Error): void
  {
    console.log("Error! " + JSON.stringify(thrownError));
    this.informationText = "Error! " + JSON.stringify(thrownError["error"]);
  }
}
