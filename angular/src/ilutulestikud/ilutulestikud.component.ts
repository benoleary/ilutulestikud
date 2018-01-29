import { Component } from '@angular/core';
import { OnInit } from '@angular/core/src/metadata/lifecycle_hooks';
import { MatDialog } from '@angular/material';
import { IlutulestikudService } from './ilutulestikud.service';
import { Player } from './models/player.model'
import { AddPlayerDialogueComponent } from './components/addplayerdialogue.component'

@Component({
  selector: 'app-ilutulestikud',
  templateUrl: './ilutulestikud.component.html',
  styleUrls: ['ilutulestikud.component.css']
})
export class IlutulestikudComponent implements OnInit
{
  selectedPlayer: Player;
  registeredPlayers: Player[];
  isAddPlayerDialogueVisible: boolean;
  newPlayerName: string;
  informationText: string;
  

  constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
  {
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.isAddPlayerDialogueVisible = false;
    this.newPlayerName = null;
    this.informationText = null;
  }

  ngOnInit(): void
  {
    this.ilutulestikudService.registeredPlayers().subscribe(
      fetchedPlayersObject => this.parsePlayers(fetchedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }
  
  parsePlayers(fetchedPlayersObject: Object): void
  {
    this.registeredPlayers.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const playerObject of fetchedPlayersObject["Players"]) {
      this.registeredPlayers.push(new Player(playerObject["Name"], playerObject["Color"]));
    }
  }

  openAddPlayerDialog(): void
  {
    let dialogRef = this.materialDialog.open(AddPlayerDialogueComponent, {
      width: '250px'
    });

    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed, result = ' + result);
    });
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
      returnedPlayersObject => this.parsePlayers(returnedPlayersObject),
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
