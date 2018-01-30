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
  availableColors: string[];
  informationText: string;
  

  constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
  {
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.availableColors = [];
    this.informationText = null;
  }

  ngOnInit(): void
  {
    this.ilutulestikudService.registeredPlayers().subscribe(
      fetchedPlayersObject => this.parsePlayers(fetchedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
    this.ilutulestikudService.availableColors().subscribe(
      fetchedColorsObject => this.parseColors(fetchedColorsObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }
  
  parsePlayers(fetchedPlayersObject: Object): void
  {
    this.registeredPlayers.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const playerObject of fetchedPlayersObject["Players"])
    {
      this.registeredPlayers.push(new Player(playerObject["Name"], playerObject["Color"]));
    }
  }
  
  parseColors(fetchedColorsObject: Object): void
  {
    this.availableColors.length = 0;

    // fetchedColorsObject["Colors"] is only an "array-like object", not an array, so does not have foreach.
    for (const colorObject of fetchedColorsObject["Colors"])
    {
      this.availableColors.push(colorObject);
    }
  }

  changeChatColor(newChatColor: string): void
  {
    this.selectedPlayer.Color = newChatColor;
    this.informationText = null;
    this.ilutulestikudService.updatePlayer(this.selectedPlayer).subscribe(
      returnedPlayersObject => this.parsePlayers(returnedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }
  

  openAddPlayerDialog(): void
  {
    let dialogRef = this.materialDialog.open(AddPlayerDialogueComponent, {
      width: '250px'
    });

    dialogRef.afterClosed().subscribe(resultFromClose =>
    {
      if (resultFromClose)
      {
        this.informationText = null;
        this.ilutulestikudService.newPlayer(resultFromClose).subscribe(
          returnedPlayersObject => this.parsePlayers(returnedPlayersObject),
          thrownError => this.handleError(thrownError),
          () => {});
      }
    });
  }

  handleError(thrownError: Error): void
  {
    console.log("Error! " + JSON.stringify(thrownError));
    this.informationText = "Error! " + JSON.stringify(thrownError["error"]);
  }
}
