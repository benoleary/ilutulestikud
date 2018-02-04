import { Component } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { MatDialog } from '@angular/material';
import { Observable } from 'rxjs/Rx';
import { IlutulestikudService } from './ilutulestikud.service';
import { Player } from './models/player.model'
import { AddPlayerDialogueComponent } from './components/addplayerdialogue.component'
import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'app-ilutulestikud',
  templateUrl: './ilutulestikud.component.html',
  styleUrls: ['ilutulestikud.component.css']
})
export class IlutulestikudComponent implements OnInit, OnDestroy
{
  informationText: string;
  selectedPlayer: Player;
  registeredPlayers: Player[];
  availableColors: string[];
  namesOfGamesWithPlayer: string[];
  gameNamesSubscription: Subscription;
  

  constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
  {
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.availableColors = [];
    this.informationText = null;
    this.namesOfGamesWithPlayer = [];
    this.gameNamesSubscription = null;
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
    this.gameNamesSubscription = Observable.timer(1000).first().subscribe(
      () => this.refreshGames());
  }

  ngOnDestroy(): void
  {
    this.gameNamesSubscription.unsubscribe();
  }

  parsePlayers(fetchedPlayersObject: Object): void
  {
    this.registeredPlayers.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const fetchedPlayer of fetchedPlayersObject["Players"])
    {
      this.registeredPlayers.push(new Player(fetchedPlayer["Name"], fetchedPlayer["Color"]));
    }
  }
  
  parseColors(fetchedColorsObject: Object): void
  {
    this.availableColors.length = 0;

    // fetchedColorsObject["Colors"] is only an "array-like object", not an array, so does not have foreach.
    for (const chatColor of fetchedColorsObject["Colors"])
    {
      this.availableColors.push(chatColor);
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

  refreshGames(): void
  {
    if (this.selectedPlayer)
    {
      this.ilutulestikudService.gamesWithPlayer(this.selectedPlayer.Name).subscribe(
        fetchedGamesObject => this.parseGames(fetchedGamesObject),
        thrownError => this.handleError(thrownError),
        () => {});
    }
  }

  parseGames(fetchedGamesObject: Object): void
  {
    this.namesOfGamesWithPlayer.length = 0;

    // fetchedGamesObject["Colors"] is only an "array-like object", not an array, so does not have foreach.
    for (const gameName of fetchedGamesObject["Games"])
    {
      this.namesOfGamesWithPlayer.push(gameName);
    }
  }

  handleError(thrownError: Error): void
  {
    console.log("Error! " + JSON.stringify(thrownError));
    this.informationText = "Error! " + JSON.stringify(thrownError["error"]);
  }
}
