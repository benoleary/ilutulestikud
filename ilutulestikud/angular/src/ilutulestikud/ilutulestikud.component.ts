import { Component } from '@angular/core';
import { OnInit } from '@angular/core/src/metadata/lifecycle_hooks';
import { IlutulestikudService } from './ilutulestikud.service';

@Component({
  selector: 'app-ilutulestikud',
  templateUrl: './ilutulestikud.component.html'
})
export class IlutulestikudComponent implements OnInit
{
  ilutulestikudService: IlutulestikudService;
  registeredPlayerNames: string[];
  isAddPlayerDialogueVisible: boolean;
  newPlayerName: string;
  informationText: string;
  playerPanelText: string;
  

  constructor(ilutulestikudService: IlutulestikudService)
  {
    this.ilutulestikudService = ilutulestikudService;
    this.isAddPlayerDialogueVisible = false;
    this.informationText = null;
    this.playerPanelText = null;
  }

  ngOnInit(): void
  {
    this.ilutulestikudService.registeredPlayerNames().subscribe(
      fetchedPlayerNamesObject => this.registeredPlayerNames = fetchedPlayerNamesObject["Names"].slice(),
      thrownError => this.informationText = "Error! " + thrownError,
      () => {});
      this.playerPanelText = "Player: [no player selected]";
  }

  availablePlayerText(): string
  {
    if (!this.registeredPlayerNames || (this.registeredPlayerNames.length == 0)) {
      return "No player names have been registered yet.";
    }

    return "Registered player names: " + this.registeredPlayerNames;
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
      returnedPlayerNamesObject =>
      {
        this.registeredPlayerNames = returnedPlayerNamesObject["Names"].slice()
      },
      thrownError =>
      {
        console.log("Error! " + JSON.stringify(thrownError));
        this.informationText = "Error! " + JSON.stringify(thrownError["error"])
      },
      () => {});
    this.isAddPlayerDialogueVisible = false;
  }
}
